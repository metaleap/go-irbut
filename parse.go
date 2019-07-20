package irbut

import (
	"strings"
)

const SrcPrelude = `

feed f1 f2:
	this.f1.f2

keep f1 f2:
	[this.f1 this].f2

call addr f:
	this.f.![@1 @addr]

//   *[this *[ [ifTrue ifFalse] 0 *[[2 3] 0 *[this 4 4 boolish]]]]
case boolish ifTrue ifFalse:
	this.foo
	foo: [ifTrue ifFalse].@addr
	addr: [2 3].@this.++boolish
	`

func Parse(src string) Noun {
	// strip top-level-only comment lines first
	for pos := strings.Index(src, "\n//"); pos >= 0; pos = strings.Index(src, "\n//") {
		if p2 := strings.IndexByte(src[pos+1:], '\n'); p2 < 0 {
			src = src[:pos+1]
		} else {
			src = src[:pos+1] + src[pos+1+p2+1:]
		}
	}

	srctopchunks := strSplitAndTrim(strings.TrimSpace(src), "\n\n", true)
	// scan all names first so earlier defs can ref to later ones
	type DefRaw []string // in order: 1 name, 0-or-more arg-names, 1 body-src
	alldefs, alldefnames := make([]DefRaw, 0, len(srctopchunks)), make(map[string]int, len(srctopchunks))
	for _, srcdef := range srctopchunks {
		srcdefhead, srcdefbody := strBreakAndTrim(srcdef, ':')
		srcdefnames := strSplitAndTrim(srcdefhead, " ", true)
		alldefnames[srcdefnames[0]], alldefs =
			len(alldefs), append(alldefs, DefRaw(append(srcdefnames, srcdefbody)))
	}
	return nil
}

func strBreakAndTrim(s string, sep byte) (left string, right string) {
	if pos := strings.IndexByte(s, sep); pos <= 0 {
		panic("expected `" + string(sep) + "` in: " + s)
	} else if left, right = strings.TrimSpace(s[:pos]), strings.TrimSpace(s[pos+1:]); left == "" {
		panic("expected something before `" + string(sep) + "` in: " + s)
	} else if right == "" {
		panic("expected something after `" + string(sep) + "` in: " + s)
	}
	return
}

func strSplitAndTrim(s string, sep string, dropEmpties bool) (r []string) {
	if len(s) != 0 {
		r = strings.Split(s, sep)
		for i := range r {
			r[i] = strings.TrimSpace(r[i])
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
