package irbut

import (
	"strconv"
)

const (
	Nil   NounAtom = 0xffffffffffffffff
	True  NounAtom = 0
	False NounAtom = 1
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

func (me *NounCell) String() string { return "[" + me.L.String() + " " + me.R.String() + "]" }

const (
	OP_AT NounAtom = iota
	OP_CONST
	OP_EVAL
	OP_ISCELL
	OP_INCR
	OP_EQ
	_ // 6 must remain free because globals (custom defs) have tree addresses 8-2 (=6), 16-2, 32-2, 64-2, 128-2 and so on
	OP_CASE
	_
	_
	_
	OP_HINT
)

// legible short-hand constructor for cells
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
	if addratom, isaddratom := addr.(NounAtom); isaddratom && addratom > 0 {
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
		} else if n, ok := at(addratom/2, tree); ok {
			return at(2+(addratom%2), n)
		}
	}
	return nil, false
}

type Prog struct {
	Globals       *NounCell
	globalsByAddr map[NounAtom]*NounCell

	OnHintStatic  func(subj Noun, discard Noun, args Noun) Noun
	OnHintDynamic func(subj Noun, discardValue Noun, discardResult Noun, args Noun) Noun

	onHintStatic  bool
	onHintDynamic bool
}

// Interp returns a `Noun` other than `code`, or `panic`s with an offending `Noun`.
func (me *Prog) Interp(code Noun) Noun {
	me.onHintDynamic, me.onHintStatic = (me.OnHintDynamic != nil), (me.OnHintStatic != nil)
	return me.interp(code)
}

func (me *Prog) interp(code Noun) Noun {
	// many of the type assertions wouldn't be necessary if not for the fact
	// that we require all `panic`s to be `Noun`-typed, to signal
	// infinite-loop-preemption aka. no-further-reducability aka. termination.
	if sf, codeiscell := code.(*NounCell); codeiscell {
		subj := sf.L
		if formula, isformulacell := sf.R.(*NounCell); isformulacell {
			op, args := formula.L, formula.R
			if _, isopcell := op.(*NounCell); isopcell {
				return ___(
					me.interp(___(subj, op)),
					me.interp(___(subj, args)),
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
						return me.interp(___(
							me.interp(___(subj, argscell.L)),
							me.interp(___(subj, argscell.R)),
						))
					}
				case OP_ISCELL:
					v := me.interp(___(subj, args))
					if _, ok := v.(*NounCell); ok {
						return True
					} else if _, ok = v.(NounAtom); ok {
						return False
					}
				case OP_INCR:
					v := me.interp(___(subj, args))
					if vatom, isvatom := v.(NounAtom); isvatom {
						return vatom + 1
					}
				case OP_EQ:
					v := me.interp(___(subj, args))
					if vcell, isvcell := v.(*NounCell); isvcell {
						if eq(vcell.L, vcell.R) {
							return True
						}
						return False
					}
				case OP_CASE:
					if isargscell {
						if boolish, isboolish := argscell.L.(NounAtom); isboolish {
							if branch, iscell := argscell.R.(*NounCell); iscell {
								if boolish == True {
									return branch.L
								} else if boolish == False {
									return branch.R
								}
							}
						}
					}
				case OP_HINT:
					if dyn, isdyn := argscell.L.(*NounCell); !isdyn {
						if me.onHintStatic {
							if n := me.OnHintStatic(subj, argscell.L, argscell.R); n != nil {
								return n
							}
						}
						return me.interp(___(subj, argscell.R))
					} else {
						if discardresult := me.interp(___(subj, dyn.R)); me.onHintDynamic {
							if n := me.OnHintDynamic(subj, argscell.L, discardresult, argscell.R); n != nil {
								return n
							}
						}
						return me.interp(___(subj, argscell.R))
					}
				default:
					if deftree := me.globalsByAddr[opcode]; deftree != nil {
						if defbagnbody, _ := deftree.L.(*NounCell); defbagnbody != nil {
							return me.interp(___(
								___(___(subj, ___(args, defbagnbody.L)), me.Globals.R),
								defbagnbody.R))
						}
					}
				}
			}
		}
	}
	panic(code)
}
