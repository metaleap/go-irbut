package nock

// OpsOrigs, if `true`, uses op-code reductions from the
// original Urbit white-paper instead of "Nock 4K" spec.
var OpsOrigs bool

const (
	OP_NOUN_AT_TREEADDR NounAtom = iota
	OP_CONST
	OP_EVALUATE
	OP_AXIOMATIC_ISCELL
	OP_AXIOMATIC_INCREMENT
	OP_AXIOMATIC_EQ
	OP_MACRO_IFTHENELSE
	OP_MACRO_COMPOSE
	OP_MACRO_EXTEND
	OP_MACRO_INVOKE
	OP_MACRO_EDIT
	OP_MACRO_HINT
)

func (NounAtom) IsCell() Noun {
	return False
}

func (*NounCell) IsCell() Noun {
	return True
}

func (me NounAtom) Increment() Noun {
	return NounAtom(me + 1)
}

func (me *NounCell) Increment() Noun {
	panic(me)
}

func (me NounAtom) eq(cmp Noun) bool {
	na, ok := cmp.(NounAtom)
	return ok && me == na
}

func (me NounAtom) Eq() Noun {
	panic(me)
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

func (me NounAtom) TreeAddr() Noun {
	panic(me)
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
	panic(me)
}

func (me NounAtom) Edit() Noun {
	panic(me)
}

func (me *NounCell) Edit() Noun {
	addr, valdst := me.L.(NounAtom), me.R.(*NounCell)
	if addr == 1 { //?                                                       #[1 a b]
		return valdst.L //>                                                  a
	}
	a, m := addr/2, addr%2
	if m == 0 { //?                                                          #[(a + a) b c]
		return nc3(a, //>                                                    #[a [b /[(a + a + 1) c]] c]
			nc(valdst.L, nan(a+a+1, valdst.R).TreeAddr()),
			valdst.R).Edit()
	} else { //?                                                             #[(a + a + 1) b c]
		return nc3(a, //>                                                    #[a [/[(a + a) c] b] c]
			nc(nc(a+a, valdst.R).TreeAddr(), valdst.L),
			valdst.R).Edit()
	}
	panic(me)
}

func (me NounAtom) Interp() Noun {
	panic(me)
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

	fr, _ := f.R.(*NounCell)
	switch code {
	case OP_NOUN_AT_TREEADDR: //?                                            *[a 0 b]
		return nc(f.R, a).TreeAddr() //>                                     /[b a]
	case OP_CONST: //?                                                       *[a 1 b]
		return f.R //>                                                       b
	case OP_AXIOMATIC_ISCELL: //?                                            *[a 3 b]
		return nc(a, f.R).Interp().IsCell() //>                              ?*[a b]
	case OP_AXIOMATIC_INCREMENT: //?                                         *[a 4 b]
		return nc(a, f.R).Interp().Increment() //>                           +*[a b]
	case OP_AXIOMATIC_EQ:
		if OpsOrigs { //?                                                    *[a 5 b]
			return nc(a, f.R).Interp().Eq() //>                              =*[a b]
		} //?                                                                *[a 5 b c]
		return nc( //>                                                       =[*[a b] *[a c]]
			nc(a, fr.L).Interp(),
			nc(a, fr.R).Interp(),
		).Eq()
	case OP_EVALUATE: //?                                                    *[a 2 b c]
		return nc( //>                                                       *[*[a b] *[a c]]
			nc(a, fr.L).Interp(),
			nc(a, fr.R).Interp(),
		).Interp()
	case OP_MACRO_IFTHENELSE: //?                                            *[a 6 b c d]
		b, cd := fr.L, fr.R.(*NounCell)
		if OpsOrigs { //>                                                    *[a 2 [0 1] 2 [1 c d] [1 0] 2 [1 2 3] [1 0] 4 4 b]
			return N(a, 2, na2(0, 1), 2, nc3(A(1), cd.L, cd.R), na2(1, 0), 2, na3(1, 2, 3), na2(1, 0), 4, 4, b).Interp()
		}
		return nc(a, nc3( //>                                                *[a *[[c d] 0 *[[2 3] 0 *[a 4 4 b]]]]
			cd, A(0), nc3(na2(2, 3), A(0), nc(a, nc3(A(4), A(4), b)).Interp()).Interp(),
		).Interp()).Interp()
	case OP_MACRO_COMPOSE: //?                                               *[a 7 b c]
		if OpsOrigs {
			return N(a, 2, fr.L, 1, fr.R).Interp() //>                       *[a 2 b 1 c]
		}
		return nc(nc(a, fr.L).Interp(), fr.R).Interp() //>                   *[*[a b] c]
	case OP_MACRO_EXTEND: //?                                                *[a 8 b c]
		if OpsOrigs {
			return N(a, 7, nc3( //>                                          *[a 7 [[7 [0 1] b] 0 1] c]
				nc3(A(7), na2(0, 1), fr.L), A(0), A(1),
			), fr.R).Interp()
		}
		return nc(nc(nc(a, fr.L).Interp(), a), fr.R).Interp() //>            *[[*[a b] a] c]
	case OP_MACRO_INVOKE: //?                                                *[a 9 b c]
		if OpsOrigs {
			return N(a, 7, fr.R, 2, na2(0, 1), 0, fr.L).Interp() //>         *[a 7 c 2 [0 1] 0 b]
		}
		return N(nc(a, fr.R).Interp(), //>                                   *[*[a c] 2 [0 1] 0 b]
			A(2), na2(0, 1), A(0), fr.L).Interp()
	case OP_MACRO_EDIT: //?                                                  *[a 10 [b c] d]
		bc, d := fr.L.(*NounCell), fr.R
		return nc3(bc.L, nc(a, bc.R).Interp(), nc(a, d).Interp()).Edit() //> #[b *[a c] *[a d]]
	case OP_MACRO_HINT:
		if frl, ok := fr.L.(*NounCell); !ok /* static hint */ { //?          *[a 11 b c]
			return nc(a, fr.R).Interp() //>                                  *[a c]
		} else /* dynamic hint */ { //?                                      *[a 11 [b c] d]
			if OpsOrigs {
				return N(a, 8, frl.R, 7, na2(0, 3), fr.R).Interp() //>       *[a 8 c 7 [0 3] d]
			}
			return nc3(nc( //>                                               *[[*[a c] *[a d]] 0 3]
				nc(a, frl.R).Interp(), nc(a, fr.R).Interp(),
			), A(0), A(3)).Interp()
		}
	}
	panic(me)
}
