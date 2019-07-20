package irbut

import (
	"strconv"
	"strings"
)

func ParseProg(src string, entryPointDefName string) *Prog {
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
		srcdefhead, srcdefbody := strBreakAndTrim(srcdef, ':', true)
		srcdefnames := strSplitAndTrim(srcdefhead, " ", true)
		if len(srcdefnames) == 0 {
			panic("expected name before `:` in: " + srcdef)
		}
		alldefnames[srcdefnames[0]], alldefs =
			len(alldefs), append(alldefs, DefRaw(append(srcdefnames, srcdefbody)))
	}

	// the prog-tree: L is entry-point addr, R is tree of (L: first def, R: `Nil` or next such sub-tree)
	progtree, defaddrs, defsbyaddr := &NounCell{L: Nil, R: Nil}, make(map[string]NounAtom, len(alldefs)), make(map[NounAtom]*NounCell, len(alldefs))
	prevtree, addr := progtree, NounAtom(8)
	// collect addrs first so each def-parse below has all globals' addrs at hand
	for i := range alldefs {
		defaddr, defname := addr-2, alldefs[i][0]
		addr = addr + addr
		if _, exists := defaddrs[defname]; exists {
			panic("duplicate global def name `" + defname + "`")
		}
		if defaddrs[defname] = defaddr; defname == entryPointDefName {
			progtree.L = defaddr
		}
	}
	// parse the defs, put them into the prog-tree
	for i := range alldefs {
		deftree := ___(parseGlobalDef(defaddrs, alldefs[i]), Nil)
		defsbyaddr[defaddrs[alldefs[i][0]]] = deftree
		prevtree.R, prevtree = deftree, deftree
	}

	return &Prog{Globals: progtree, globalsByAddr: defsbyaddr}
}

func parseGlobalDef(globalDefs map[string]NounAtom, nameArgsBody []string) *NounCell {
	args := make(map[string]NounAtom, len(nameArgsBody)-2)
	for i := 1; i < len(nameArgsBody)-1; i++ {
		argname := nameArgsBody[i]
		if _, exists := args[argname]; exists {
			panic("duplicate arg name `" + argname + "` in global def `" + nameArgsBody[0] + "`")
		}
		args[argname] = Nil
	}

	srclines := strSplitAndTrim(nameArgsBody[len(nameArgsBody)-1], "\n", true)
	if len(srclines) == 0 {
		panic("expected body following `:` for def `" + nameArgsBody[0] + "`")
	}
	localdefaddrs, localstree := make(map[string]NounAtom, len(srclines)-1), &NounCell{L: Nil, R: Nil}
	if len(srclines) > 1 { // we have named LOCAL DEFS
		type todo struct {
			name    string
			bodySrc string
			addr    NounAtom
			subTree *NounCell
		}
		prevtree, addr, locals := localstree, NounAtom(8), make([]todo, len(srclines)-1)
		for i := 1; i < len(srclines); i++ {
			localsrc, def := srclines[i], &locals[i-1]
			if def.name, def.bodySrc = strBreakAndTrim(localsrc, ':', true); def.name == "" {
				panic("expected name preceding `:` for local def in: " + localsrc)
			} else if def.bodySrc == "" {
				panic("expected body following `:` for local def in: " + localsrc)
			} else {
				nexttree := &NounCell{L: Nil, R: Nil}
				def.subTree, def.addr, prevtree.L, prevtree.R = prevtree, addr-2, nil, nexttree
				prevtree, addr = nexttree, addr+addr
			}
		}
		for i := range locals {
			def := &locals[i]
			if _, exists := localdefaddrs[def.name]; exists {
				panic("duplicate local def name `" + def.name + "`")
			}
			localdefaddrs[def.name] = def.addr
			def.subTree.L = parseExpr(nameArgsBody[0], def.name, globalDefs, args, localdefaddrs, def.bodySrc)
		}
	}

	// dealt with the locals, now parse the def's body expr
	return ___(
		localstree,
		parseExpr(nameArgsBody[0], "", globalDefs, args, localdefaddrs, srclines[0]),
	)
}

func parseExpr(curGlobalDefName string, curLocalDefName string, globalDefs map[string]NounAtom, args map[string]NounAtom, localDefs map[string]NounAtom, src string) (expr Noun) {
	toks := strTokens(src)
	for _, tok := range toks {
		var cur Noun
		if tok[0] >= '0' && tok[0] <= '9' {
			if ui, e := strconv.ParseUint(tok, 10, 64); e != nil {
				panic(e)
			} else {
				cur = NounAtom(ui)
			}
		}
		if expr == nil {
			expr = cur
		}
	}

	if expr == nil {
		panic("expression expected in: " + src)
	}
	return
}

func ParseExpr(src string) Noun {
	return parseExpr("", "", nil, nil, nil, src)
}

func strBreakAndTrim(s string, sep byte, stripComments bool) (left string, right string) {
	if pos := strings.IndexByte(s, sep); pos <= 0 {
		panic("expected `" + string(sep) + "` in: " + s)
	} else if left, right = strings.TrimSpace(strStripCommentIf(stripComments, s[:pos])), strings.TrimSpace(strStripCommentIf(stripComments, s[pos+1:])); left == "" {
		panic("expected something preceding `" + string(sep) + "` in: " + s)
	} else if right == "" {
		panic("expected something following `" + string(sep) + "` in: " + s)
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

func strTokens(src string) (toks []string) {
	inbracketsince, inwordsince := -1, -1
	var brackets int
	for i := 0; i < len(src); i++ {
		if src[i] == '[' {
			if brackets++; inbracketsince == -1 {
				if inbracketsince = i; inwordsince != -1 {
					inwordsince, toks = -1, append(toks, src[inwordsince:i])
				}
			}
		}
		if brackets == 0 {
			isalphanum := (src[i] >= '0' && src[i] <= '9') || (src[i] >= 'A' && src[i] <= 'Z') || (src[i] >= 'a' && src[i] <= 'z')
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
			if brackets--; brackets == 0 {
				inbracketsince, toks = -1, append(toks, src[inbracketsince:i+1])
			}
		}
	}
	if inwordsince != -1 {
		toks = append(toks, src[inwordsince:])
	}
	return
}
