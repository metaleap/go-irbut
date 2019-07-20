package irbut

import (
	"strconv"
	"strings"
)

func ParseProg(src string) *Prog {
	// strip top-level-only comment lines first
	if strings.HasPrefix(src, "//") {
		src = "\n" + src
	}
	for pos := strings.Index(src, "\n//"); pos >= 0; pos = strings.Index(src, "\n//") {
		if p2 := strings.IndexByte(src[pos+1:], '\n'); p2 < 0 {
			src = src[:pos]
		} else {
			src = src[:pos] + src[pos+1+p2:]
		}
	}

	src = strings.TrimSpace(src)
	srctopchunks := strSplitAndTrim(src, "\n\n", true)
	// scan all names first so earlier defs can ref to later ones
	type DefRaw []string // in order: 1 name, 0-or-more arg-names, 1 body-src
	alldefs, alldefnames := make([]DefRaw, 0, len(srctopchunks)), make(map[string]int, len(srctopchunks))
	for _, srcdef := range srctopchunks {
		srcdefhead, srcdefbody := strBreakAndTrim(srcdef, ':', false, "?")
		srcdefnames := strSplitAndTrim(srcdefhead, " ", true)
		if len(srcdefnames) == 0 {
			panic("expected name before `:` near: " + srcdef)
		}
		alldefnames[srcdefnames[0]], alldefs =
			len(alldefs), append(alldefs, DefRaw(append(srcdefnames, srcdefbody)))
	}

	// the prog-tree: L is first def or `None`, R is `None` or another such sub-tree
	progtree, defaddrs := &NounCell{L: None, R: None}, make(map[string]NounAtom, len(alldefs))
	prevtree, addr := progtree, NounAtom(4)
	// collect addrs first so each def-parse below has all globals' addrs at hand
	for i := range alldefs {
		defaddr, defname := addr-2, alldefs[i][0]
		if _, exists := defaddrs[defname]; exists {
			panic("duplicate global def name `" + defname + "`")
		}
		defaddrs[defname], addr = defaddr, addr+addr
	}
	defsbyaddr := make(map[NounAtom]*NounCell, len(alldefs))
	// parse the defs, put them into the prog-tree
	for i := range alldefs {
		def, nexttree := parseGlobalDef(defaddrs, alldefs[i]), &NounCell{L: None, R: None}
		defsbyaddr[defaddrs[alldefs[i][0]]], prevtree.L, prevtree.R =
			def, def, nexttree
		prevtree = nexttree
	}

	return &Prog{Globals: progtree, globalsByAddr: defsbyaddr}
}

func parseGlobalDef(globalDefs map[string]NounAtom, nameArgsBody []string) *NounCell {
	ctx := ctxParse{
		argAddrs:         make(map[string]NounAtom, len(nameArgsBody)-2),
		globalDefAddrs:   globalDefs,
		curGlobalDefName: nameArgsBody[0],
	}
	for i := 1; i < len(nameArgsBody)-1; i++ {
		argname := nameArgsBody[i]
		if _, exists := ctx.argAddrs[argname]; exists {
			panic("in `" + nameArgsBody[0] + "`: duplicate arg name `" + argname + "`")
		}
		ctx.argAddrs[argname] = None
	}

	srclines := strSplitAndTrim(nameArgsBody[len(nameArgsBody)-1], "\n", true)
	if len(srclines) == 0 {
		panic("in `" + nameArgsBody[0] + "`: expected body following `:`")
	}
	localstree := &NounCell{L: None, R: None}
	if ctx.localDefAddrs = make(map[string]NounAtom, len(srclines)-1); len(srclines) > 1 {
		type todo struct {
			name    string
			bodySrc string
			addr    NounAtom
			subTree *NounCell
		}
		prevtree, addr, locals := localstree, NounAtom(8), make([]todo, len(srclines)-1)
		for i := 1; i < len(srclines); i++ {
			localsrc, def := srclines[i], &locals[i-1]
			if def.name, def.bodySrc = strBreakAndTrim(localsrc, ':', true, nameArgsBody[0]); def.name == "" {
				panic("in `" + nameArgsBody[0] + "`: expected name preceding `:` for local def near: " + localsrc)
			} else if def.bodySrc == "" {
				panic("in `" + strJoin2(nameArgsBody[0], "/", def.name) + "`: expected body following `:` for local def near: " + localsrc)
			} else {
				nexttree := &NounCell{L: None, R: None}
				def.subTree, def.addr, prevtree.L, prevtree.R = prevtree, addr-2, nil, nexttree
				prevtree, addr = nexttree, addr+addr

				if _, exists := ctx.localDefAddrs[def.name]; exists {
					panic("in `" + nameArgsBody[0] + "`: duplicate local def name `" + def.name + "`")
				}
				ctx.localDefAddrs[def.name] = def.addr
			}
		}
		for i := range locals {
			def := &locals[i]
			ctx.curLocalDefName = def.name
			def.subTree.L = ctx.parseExpr(def.bodySrc)
		}
	}

	// dealt with the locals, now parse the def's body expr
	ctx.curLocalDefName = ""
	return ___(
		localstree,
		ctx.parseExpr(srclines[0]),
	)
}

type ctxParse struct {
	curGlobalDefName string
	curLocalDefName  string
	globalDefAddrs   map[string]NounAtom
	argAddrs         map[string]NounAtom
	localDefAddrs    map[string]NounAtom
}

func (me *ctxParse) parseExpr(src string) (expr Noun) {
	fail := func(tok string, msg string) {
		panic("in `" + strJoin2(me.curGlobalDefName, "/", me.curLocalDefName) + "` near `" + tok + "`: " + msg)
	}
	toks, numopenbrackets := strTokens(src)
	if numopenbrackets != 0 {
		fail(toks[len(toks)-1], strconv.FormatInt(int64(numopenbrackets), 10)+" unclosed bracket(s)")
	}

	for _, tok := range toks {
		var cur Noun

		if tok[0] == '[' {
			if cur = me.parseCell(strSplitAndTrim(tok[1:len(tok)-1], " ", true)); cur == nil {
				fail(tok, "expected at least 2 cell elements")
			}

		} else if tok[0] >= '0' && tok[0] <= '9' {
			if ui, e := strconv.ParseUint(tok, 0, 64); e != nil {
				fail(tok, e.Error())
			} else {
				cur = NounAtom(ui)
			}

		} else if tok[0] == '_' || (tok[0] >= 'A' && tok[0] <= 'Z') || (tok[0] >= 'a' && tok[0] <= 'z') {
			addr := me.argAddrs[tok]
			if addr == 0 {
				if addr = me.localDefAddrs[tok]; addr == 0 {
					if addr = me.globalDefAddrs[tok]; addr == 0 {
						switch tok {
						case "this":
							addr = None
						default:
							fail(tok, "unknown name")
						}
					}
				}
			}
			cur = addr
		} else if len(tok) == 1 {
			switch tok[0] {
			case ':':
				fail(tok, "wrong line for a local def (`:` is not permissible in expressions)")
			case '.':
				cur = None // temp, TODO
			default:
				cur = NounAtom(tok[0])
			}
		}
		if cur == nil {
			fail(tok, "unrecognized token")
		} else if expr == nil {
			expr = cur
		}
	}

	if expr == nil {
		fail(src, "expression expected")
	}
	return
}

func (me *ctxParse) parseCell(src []string) *NounCell {
	if len(src) > 1 {
		var cell NounCell
		cell.L = me.parseExpr(src[0])
		if len(src) == 2 {
			cell.R = me.parseExpr(src[1])
		} else {
			cell.R = me.parseCell(src[1:])
		}
		return &cell
	}
	return nil
}

func ParseExpr(src string) Noun {
	ctx := ctxParse{
		curGlobalDefName: "<input>",
		globalDefAddrs:   make(map[string]NounAtom, 0),
		argAddrs:         make(map[string]NounAtom, 0),
		localDefAddrs:    make(map[string]NounAtom, 0),
	}
	return ctx.parseExpr(src)
}

func strBreakAndTrim(s string, sep byte, stripComments bool, nameForErrs string) (left string, right string) {
	if pos := strings.IndexByte(s, sep); pos <= 0 {
		panic("in `" + nameForErrs + "`: expected `" + string(sep) + "` near: " + s)
	} else if left, right = strings.TrimSpace(strStripCommentIf(stripComments, s[:pos])), strings.TrimSpace(strStripCommentIf(stripComments, s[pos+1:])); left == "" {
		panic("in `" + nameForErrs + "`: expected token(s) preceding `" + string(sep) + "` near: " + s)
	} else if right == "" {
		panic("in `" + nameForErrs + "`: expected token(s) following `" + string(sep) + "` near: " + s)
	}
	return
}

func strSplitAndTrim(s string, sep string, dropEmpties bool) (r []string) {
	if len(s) != 0 {
		r = strings.Split(s, sep)
		for i := range r {
			r[i] = strings.TrimSpace(r[i])
		}
		if sep == "\n" {
			for i := range r {
				r[i] = strStripCommentIf(true, r[i])
			}
		}
		if dropEmpties {
			for i := 0; i < len(r); i++ {
				if r[i] == "" {
					r = append(r[:i], r[i+1:]...)
					i--
				}
			}
		}
	}
	return
}

func strStripCommentIf(when bool, s string) string {
	if when {
		if pos := strings.Index(s, "//"); pos >= 0 {
			return s[:pos]
		}
	}
	return s
}

func strTokens(src string) (toks []string, numOpenBrackets int) {
	inbracketsince, inwordsince := -1, -1
	for i := 0; i < len(src); i++ {
		if src[i] == '[' {
			if numOpenBrackets++; inbracketsince == -1 {
				if inbracketsince = i; inwordsince != -1 {
					inwordsince, toks = -1, append(toks, src[inwordsince:i])
				}
			}
		}
		if numOpenBrackets == 0 {
			isalphanum := src[i] == '_' || (src[i] >= '0' && src[i] <= '9') || (src[i] >= 'A' && src[i] <= 'Z') || (src[i] >= 'a' && src[i] <= 'z')
			if !isalphanum {
				if inwordsince != -1 {
					inwordsince, toks = -1, append(toks, src[inwordsince:i])
				}
				if src[i] != ' ' && src[i] != '\t' {
					toks = append(toks, string(src[i]))
				}
			} else if inwordsince == -1 {
				inwordsince = i
			}
		} else if src[i] == ']' {
			if numOpenBrackets--; numOpenBrackets == 0 {
				inbracketsince, toks = -1, append(toks, src[inbracketsince:i+1])
			}
		}
	}
	if inbracketsince != -1 {
		toks = append(toks, src[inbracketsince:])
	} else if inwordsince != -1 {
		toks = append(toks, src[inwordsince:])
	}
	return
}

func strJoin2(s1 string, sep string, s2 string) string {
	if s1 == "" {
		return s2
	} else if s2 == "" {
		return s1
	} else {
		return s1 + sep + s2
	}
}
