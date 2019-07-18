package nock

const ui32Max uint32 = 0xFFFFFFFF
const ui64Max uint64 = 0xFFFFFFFFFFFFFFFF

const True NounAtom = 0
const False NounAtom = 1

type Noun interface {
	DepthTest() Noun
	Increment() Noun
	eq(Noun) bool
	Eq() Noun
	TreeAddr() Noun
}

type tAtom = uint32

type NounAtom tAtom

type NounCell struct {
	L Noun
	R Noun
}

func N(v ...interface{}) Noun {
	if l := len(v); l > 1 {
		ret := &NounCell{N(v[l-2]), N(v[l-1])}
		for i := 3; i <= l; i++ {
			ret = &NounCell{N(v[l-i]), ret}
		}
		return ret
	} else {
		return v[0].(NounAtom)
	}
}
