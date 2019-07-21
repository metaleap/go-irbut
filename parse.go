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
	// next, scan all names only before actual parsing so earlier defs can ref to later ones
	type DefRaw []string // in order: 1 name, 0-or-more arg-names, 1 body-src
	alldefs := make([]DefRaw, 0, len(srctopchunks))
	for _, srcdef := range srctopchunks {
		srcdefhead, srcdefbody := strBreakAndTrim(srcdef, ':', false, "?")
		srcdefnameandargs := strSplitAndTrim(srcdefhead, " ", true)
		if len(srcdefnameandargs) == 0 {
			panic("expected name before `:` near: " + srcdef)
		}
		alldefs = append(alldefs, DefRaw(append(srcdefnameandargs, srcdefbody)))
	}

	prog, defidxs := Prog{Globals: make([]*NounCell, 0, len(alldefs))}, make(map[string]NounAtom, len(alldefs))
	// collect idxs first so each def-parse below has all globals at hand
	for _, defraw := range alldefs {
		defname, idx := defraw[0], NounAtom(len(prog.Globals))
		if _, exists := defidxs[defname]; exists {
			panic("duplicate global def name `" + defname + "`")
		}
		for idx > 32 && idx < 127 &&
			(idx == OP_AT || idx == OP_CASE || idx == OP_CONST || idx == OP_EQ || idx == OP_EVAL || idx == OP_HINT || idx == OP_INCR || idx == OP_ISCELL) {
			idx, prog.Globals = idx+1, append(prog.Globals, nil)
		}
		defidxs[defname], prog.Globals = idx, append(prog.Globals, &NounCell{})
	}
	// parse the defs, put them into the prog-tree
	for _, defraw := range alldefs {
		ctxdef := ctxParseGlobalDef{
			curGlobalDefName: defraw[0],
			globalDefs:       defidxs,
		}
		ctxdef.parse(prog.Globals[defidxs[ctxdef.curGlobalDefName]], defraw)
	}

	return &prog
}

type ctxParseGlobalDef struct {
	curGlobalDefName string
	curLocalDefName  string
	globalDefs       map[string]NounAtom
	argAddrs         map[string]NounAtom
	localDefAddrs    map[string]NounAtom
}

func (me *ctxParseGlobalDef) parse(into *NounCell, nameArgsBody []string) {
	defname := me.curGlobalDefName
	me.argAddrs = make(map[string]NounAtom, len(nameArgsBody)-2)
	for i := 1; i < len(nameArgsBody)-1; i++ {
		argname := nameArgsBody[i]
		if _, exists := me.argAddrs[argname]; exists {
			panic("in `" + defname + "`: duplicate arg name `" + argname + "`")
		}
		me.argAddrs[argname] = None
	}

	srclines := strSplitAndTrim(nameArgsBody[len(nameArgsBody)-1], "\n", true)
	if len(srclines) == 0 {
		panic("in `" + defname + "`: expected body following `:`")
	}
	localstree := &NounCell{L: None, R: None}
	if me.localDefAddrs = make(map[string]NounAtom, len(srclines)-1); len(srclines) > 1 {
		prevtree, addr, locals := localstree, NounAtom(8), make([]struct {
			name    string
			bodySrc string
			addr    NounAtom
			subTree *NounCell
		}, len(srclines)-1)
		for i := 1; i < len(srclines); i++ {
			ldef, ldefsrc := &locals[i-1], srclines[i]
			if ldef.name, ldef.bodySrc = strBreakAndTrim(ldefsrc, ':', true, defname); ldef.name == "" {
				panic("in `" + defname + "`: expected name preceding `:` for local def near: " + ldefsrc)
			} else if ldef.bodySrc == "" {
				panic("in `" + strJoin2(defname, "/", ldef.name) + "`: expected body following `:` for local def near: " + ldefsrc)
			} else {
				nexttree := &NounCell{L: None, R: None}
				ldef.subTree, ldef.addr, prevtree.L, prevtree.R = prevtree, addr-2, nil, nexttree
				prevtree, addr = nexttree, addr+addr

				if _, exists := me.localDefAddrs[ldef.name]; exists {
					panic("in `" + defname + "`: duplicate local def name `" + ldef.name + "`")
				}
				me.localDefAddrs[ldef.name] = ldef.addr
			}
		}
		for i := range locals {
			def := &locals[i]
			me.curLocalDefName = def.name
			def.subTree.L = me.parseExpr(def.bodySrc)
		}
	}

	// dealt with the locals, now parse the def's body expr
	me.curLocalDefName = ""
	into.L, into.R = localstree, me.parseExpr(srclines[0])
}

func (me *ctxParseGlobalDef) parseExpr(src string) (expr Noun) {
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
					if addrtmp := me.globalDefs[tok]; addrtmp == 0 {
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

func (me *ctxParseGlobalDef) parseCell(src []string) *NounCell {
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
	ctx := ctxParseGlobalDef{
		curGlobalDefName: "<input>",
		globalDefs:       make(map[string]NounAtom, 0),
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
