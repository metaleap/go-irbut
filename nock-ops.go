package nock

const (
	OP_NOUN_AT_TREEADDR NounAtom = iota
	OP_CONST
	OP_TURING
	OP_AXIOMATIC_DEPTHTEST
	OP_AXIOMATIC_INCREMENT
	OP_AXIOMATIC_EQ
	OP_MACRO_IFTHENELSE
	OP_MACRO_COMPOSE
	OP_MACRO_VARDECL_AKA_STACKPUSH
	OP_MACRO_PTRMETHOD
	OP_MACRO_HINT
)

func (NounAtom) DepthTest() Noun {
	return False
}

func (*NounCell) DepthTest() Noun {
	return True
}

func (me NounAtom) Increment() Noun {
	return NounAtom(me + 1)
}

func (*NounCell) Increment() Noun {
	panic("*NounCell.Increment")
}

func (me NounAtom) eq(cmp Noun) bool {
	na, ok := cmp.(NounAtom)
	return ok && me == na
}

func (NounAtom) Eq() Noun {
	panic("NounAtom.Eq")
}

func (me *NounCell) eq(cmp Noun) bool {
	nc, _ := cmp.(*NounCell)
	return (me == nc) || (me.L.eq(nc.L) && me.R.eq(nc.R))
}

func (me *NounCell) Eq() Noun {
	if me.L.eq(me.R) {
		return True
	}
	return False
}

func (NounAtom) TreeAddr() Noun {
	panic("NounAtom.TreeAddr")
}

func (me *NounCell) TreeAddr() Noun {
	head, isheadatom := me.L.(NounAtom)
	tail, istailcell := me.R.(*NounCell)

	if isheadatom {
		if head == 1 { //?                                                   /[1 a]
			return me.R //>                                                  a
		}
		if istailcell {
			if head == 2 { //?                                               /[2 a b]
				return tail.L //>                                            a
			} else if head == 3 { //?                                        /[3 a b]
				return tail.R //>                                            b
			}
		}
		m, a, b := head%2, head/2, me.R                //?                   /[(a+a+0|1) b]
		return nc(2+m, nc(a, b).TreeAddr()).TreeAddr() //>                   /[(2+0|1) /[a b]]
	}
	panic("*NounCell.TreeAddr")
}

func (NounAtom) Interp() Noun {
	panic("NounAtom.Interp")
}

func (me *NounCell) Interp() Noun {
	type A = NounAtom // just need a local short-hand below --- verbosity abounds!

	a, f := me.L, me.R.(*NounCell)
	code, iscode := f.L.(NounAtom)
	if !iscode { //?                                                         *[a [b c] d]
		return nc( //>                                                       [*[a b c] *[a d]]
			nc(a, f.L).Interp(),
			nc(a, f.R).Interp(),
		)
	}

	if code < 6 {
		switch b := f.R; code {
		case OP_NOUN_AT_TREEADDR: //?                                        *[a 0 b]
			return nc(b, a).TreeAddr() //>                                   /[b a]
		case OP_CONST: //?                                                   *[a 1 b]
			return b //>                                                     b
		case OP_AXIOMATIC_DEPTHTEST: //?                                     *[a 3 b]
			return nc(a, b).Interp().DepthTest() //>                         ?*[a b]
		case OP_AXIOMATIC_INCREMENT: //?                                     *[a 4 b]
			return nc(a, b).Interp().Increment() //>                         +*[a b]
		case OP_AXIOMATIC_EQ: //?                                            *[a 5 b]
			return nc(a, b).Interp().Eq() //>                                =*[a b]
		}
	}

	fr := f.R.(*NounCell)
	switch code {
	case OP_TURING: //?                                                      *[a 2 b c]
		return nc( //>                                                       *[*[a b] *[a c]]
			nc(a, fr.L).Interp(),
			nc(a, fr.R).Interp(),
		).Interp()
	case OP_MACRO_IFTHENELSE: //?                                            *[a 6 b c d]
		b, cd := fr.L, fr.R.(*NounCell) //>                                  *[a 2 [0 1] 2 [1 c d] [1 0] 2 [1 2 3] [1 0] 4 4 b]
		return N(a, 2, na2(0, 1), 2, nc3(A(1), cd.L, cd.R), na2(1, 0), 2, na3(1, 2, 3), na2(1, 0), 4, 4, b).Interp()
	case OP_MACRO_COMPOSE: //?                                               *[a 7 b c]
		return N(a, 2, fr.L, 1, fr.R).Interp() //>                           *[a 2 b 1 c]
	case OP_MACRO_VARDECL_AKA_STACKPUSH: //?                                 *[a 8 b c]
		return N(a, 7, nc3( //>                                              *[a 7 [[7 [0 1] b] 0 1] c]
			nc3(A(7), na2(0, 1), fr.L), A(0), A(1),
		), fr.R).Interp()
	case OP_MACRO_PTRMETHOD: //?                                             *[a 9 b c]
		return N(a, 7, fr.R, 2, na2(0, 1), 0, fr.L).Interp() //>             *[a 7 c 2 [0 1] 0 b]
	case OP_MACRO_HINT:
		if frl, ok := fr.L.(*NounCell); ok /* dynamic hint */ { //?          *[a 10 [b c] d]
			return N(a, 8, frl.R, 7, na2(0, 3), fr.R).Interp() //>           *[a 8 c 7 [0 3] d]
		} else /* static hint */ { //?                                       *[a 10 b c]
			return nc(a, fr.R).Interp() //>                                  *[a c]
		}
	}
	panic("*NounCell.Interp")
}
