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
	switch b := f.R; code {
	case 0: //?                                                              *[a 0 b]
		return nc(b, a).TreeAddr() //>                                       /[b a]
	case 1: //?                                                              *[a 1 b]
		return b //>                                                         b
	case 3: //?
	case 4: //?
	case 5: //?
	}

	panic("*NounCell.Interp")
}
