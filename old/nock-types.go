package nock

const ui32Max uint32 = 0xFFFFFFFF
const ui64Max uint64 = 0xFFFFFFFFFFFFFFFF

const True NounAtom = 0
const False NounAtom = 1

// Noun methods either return another `Noun` or `panic` with the current `Noun`.
type Noun interface {
	IsCell() Noun
	Increment() Noun
	eq(Noun) bool
	Eq() Noun
	TreeAddr() Noun
	Edit() Noun
	Interp() Noun
}

type nounAtom = uint32

type NounAtom nounAtom

type NounCell struct {
	L Noun
	R Noun
}

func nac(l NounAtom, r *NounCell) *NounCell  { return &NounCell{L: l, R: r} }
func nan(l NounAtom, r Noun) *NounCell       { return &NounCell{L: l, R: r} }
func nc(l Noun, r Noun) *NounCell            { return &NounCell{L: l, R: r} }
func nc3(l Noun, rl Noun, rr Noun) *NounCell { return &NounCell{L: l, R: &NounCell{L: rl, R: rr}} }
func na2(l NounAtom, r NounAtom) *NounCell   { return &NounCell{L: l, R: r} }
func na3(l NounAtom, rl NounAtom, rr NounAtom) *NounCell {
	return &NounCell{L: l, R: &NounCell{L: rl, R: rr}}
}

func N(v ...interface{}) Noun {
	if l := len(v); l > 1 {
		ret := &NounCell{N(v[l-2]), N(v[l-1])}
		for i := 3; i <= l; i++ {
			ret = &NounCell{N(v[l-i]), ret}
		}
		return ret
	}
	switch t := v[0].(type) {
	case NounAtom:
		return t
	case *NounCell:
		return t
	case int:
		return NounAtom(t)
	case nounAtom:
		return NounAtom(t)
	case NounCell:
		return &t
	default:
		return t.(Noun)
	}
}
