package nock

const ui32Max uint32 = 0xFFFFFFFF
const ui64Max uint64 = 0xFFFFFFFFFFFFFFFF

type Noun interface {
}

type NounAtom uint32

type NounCell struct {
	A Noun
	B Noun
}

func N(v ...interface{}) Noun {
	if l := len(v); l > 1 {
		ret := &NounCell{l - 2, l - 1}
		for i := 3; i <= l; i++ {
			ret = &NounCell{l - i, ret}
		}
		return ret
	} else {
		if i, ok1 := v[0].(int); ok1 {
			return NounAtom(i)
		} else if ui32, ok := v[0].(uint32); ok {
			return NounAtom(ui32)
		} else if ui, ok2 := v[0].(uint); ok2 {
			return NounAtom(ui)
		}
		panic(v[0])
	}
}
