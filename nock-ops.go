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
		if head == 1 { //?                                                   [1 a]
			return me.R //>                                                  a
		}
		if istailcell {
			if head == 2 { //?                                               [2 a b]
				return tail.L //>                                            a
			} else if head == 3 { //?                                        [3 a b]
				return tail.R //>                                            b
			}
		}
		m, a, b := head%2, head/2, me.R                        //?           /[(a+a+0|1) b]
		return N(NounAtom(2+m), N(a, b).TreeAddr()).TreeAddr() //>           /[(2+0|1) /[a b]]
	}
	panic("*NounCell.TreeAddr")
}
