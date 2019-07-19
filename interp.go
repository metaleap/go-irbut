package irbut

import (
	"strconv"
)

type Noun interface {
	String() string
}

type NounAtom uint64

func (me NounAtom) String() string { return strconv.FormatUint(uint64(me), 10) }

type NounCell struct {
	L Noun
	R Noun
}

func (me *NounCell) String() string { return "<" + me.L.String() + " " + me.R.String() + ">" }

const (
	True  NounAtom = 0
	False NounAtom = 1
)

const (
	OP_AT NounAtom = iota
	OP_CONST
	OP_EVAL
	OP_ISCELL
	OP_INCR
	OP_EQ
	_
	_
	_
	_
	_
	OP_HINT
)

// short-hand constructor for cells
func ___(left Noun, right Noun) *NounCell { return &NounCell{L: left, R: right} }

// structural/value equality (no identity)
func eq(noun1 Noun, noun2 Noun) bool {
	if a1, oka1 := noun1.(NounAtom); oka1 {
		if a2, oka2 := noun2.(NounAtom); oka2 {
			return a1 == a2
		}
	}
	if c1, okc1 := noun1.(*NounCell); okc1 {
		if c2, okc2 := noun2.(*NounCell); okc2 {
			return eq(c1.L, c2.L) && eq(c1.R, c2.R)
		}
	}
	return false
}

// tree-addressing scheme
func at(addr Noun, tree Noun) (Noun, bool) {
	if addratom, isaddratom := addr.(NounAtom); isaddratom {
		if addratom == 1 {
			return tree, true
		}
		if addratomis2 := (addratom == 2); addratomis2 || addratom == 3 {
			if treecell, istreecell := tree.(*NounCell); istreecell {
				if addratomis2 {
					return treecell.L, true
				}
				return treecell.R, true
			}
		} else { // *only* if greater 3!
			if n, ok := at(addratom/2, tree); ok {
				return at(2+(addratom%2), n)
			}
		}
	}
	return nil, false
}

// Interp returns a `Noun` other than `code`, or `panic`s with an offending `Noun`.
func Interp(code Noun) Noun {
	// many of the type assertions wouldn't be necessary if not for the fact
	// that we require all `panic`s to be `Noun`-typed, to signal
	// infinite-loop-preemption aka. no-further-reducability aka. termination.
	if sf, codeiscell := code.(*NounCell); codeiscell {
		subj := sf.L
		if formula, isformulacell := sf.R.(*NounCell); isformulacell {
			op, args := formula.L, formula.R
			if _, isopcell := op.(*NounCell); isopcell {
				return ___(
					Interp(___(subj, op)),
					Interp(___(subj, args)),
				)
			}
			if opcode, isopcode := op.(NounAtom); isopcode {
				argscell, isargscell := args.(*NounCell)
				switch opcode {
				case OP_AT:
					if n, ok := at(args, subj); ok {
						return n
					}
				case OP_CONST:
					return args
				case OP_EVAL:
					if isargscell {
						return Interp(___(
							Interp(___(subj, argscell.L)),
							Interp(___(subj, argscell.R)),
						))
					}
				case OP_ISCELL:
					v := Interp(___(subj, args))
					if _, ok := v.(*NounCell); ok {
						return True
					} else if _, ok = v.(NounAtom); ok {
						return False
					}
				case OP_INCR:
					v := Interp(___(subj, args))
					if vatom, isvatom := v.(NounAtom); isvatom {
						return vatom + 1
					}
				case OP_EQ:
					v := Interp(___(subj, args))
					if vcell, isvcell := v.(*NounCell); isvcell {
						if eq(vcell.L, vcell.R) {
							return True
						}
						return False
					}
					// case OP_HINT:
					// default:
					// 	return OnCustomOpCode(subj, opcode, args)
				}
			}
		}
	}
	panic(code)
}
