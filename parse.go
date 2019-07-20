package irbut

import (
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
			localdefaddrs[def.name] = def.addr
			def.subTree.L = parseExpr(globalDefs, args, localdefaddrs, def.bodySrc)
		}
	}

	// dealt with the locals, now parse the def's body expr
	return ___(
		localstree,
		parseExpr(globalDefs, args, localdefaddrs, srclines[0]),
	)
}

func parseExpr(globalDefs map[string]NounAtom, args map[string]NounAtom, localDefs map[string]NounAtom, body string) Noun {
	return nil
}

func ParseExpr(body string) Noun {
	return parseExpr(nil, nil, nil, body)
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
