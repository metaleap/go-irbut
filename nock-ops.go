package nock

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
		case 0: //?                                                              *[a 0 b]
			return nc(b, a).TreeAddr() //>                                       /[b a]
		case 1: //?                                                              *[a 1 b]
			return b //>                                                         b
		case 3: //?                                                              *[a 3 b]
			return nc(a, b).Interp().DepthTest() //                              ?*[a b]
		case 4: //?                                                              *[a 4 b]
			return nc(a, b).Interp().Increment() //                              +*[a b]
		case 5: //?
			return nc(a, b).Interp().Eq() //                                     =*[a b]
		}
	}
	fr := f.R.(*NounCell)
	switch code {
	case 2: //?                                                              *[a 2 b c]
		return nc( //>                                                       *[*[a b] *[a c]]
			nc(a, fr.L).Interp(),
			nc(a, fr.R).Interp(),
		).Interp()
	case 6:
	case 7: //?                                                              *[a 7 b c]
		return N(a, 2, fr.L, 1, fr.R).Interp() //>                           *[a 2 b 1 c]
	case 8:
	case 9:
	case 10:
		if frl, ok := fr.L.(*NounCell); ok { //?                             *[a 10 [b c] d]
			return N(a, 8, frl.R, 7, naa(0, 3), fr.R).Interp() //>           *[a 8 c 7 [0 3] d]
		} else { //?                                                         *[a 10 b c]
			return nc(a, fr.R).Interp() //>                                  *[a c]
		}
	}
	panic("*NounCell.Interp")
}
