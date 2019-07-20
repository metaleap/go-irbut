package main

import (
	"github.com/metaleap/irbut"
	"time"
)

const (
	srcSimple = `
main:
	this.>123
`
)

type noun = irbut.Noun
type ª = irbut.NounAtom
type º = irbut.NounCell

const ø = irbut.Nil

func ___(l noun, r noun) *º { return &º{L: l, R: r} }

func main() {
	out := func(n noun) { println(n.String()) }

	timestarted := time.Now()
	prog := irbut.ParseProg(irbut.SrcPrelude+"\n\n"+srcSimple, "main")
	println(time.Since(timestarted).String())

	out(prog.Interp(___(ø, ___(irbut.OP_CONST, ª(234)))))
	out(prog.Interp(___(ø, ___(irbut.OP_ISCELL, ___(irbut.OP_CONST, ª(123))))))
	out(prog.Interp(___(ø, ___(irbut.OP_ISCELL, ___(irbut.OP_CONST, ___(ª(123), ª(321)))))))
	out(prog.Interp(___(ø, ___(irbut.OP_CASE, ___(irbut.False, ___(ª(123), ª(321)))))))

	sometree := ___(___(ª(44), ª(55)), ___(ª(66), ___(ª(414), ª(515))))
	for i := 1; i < 16; i++ {
		if i < 8 || i == 14 || i == 15 {
			print("@", i, "\t->\t")
			out(prog.Interp(___(
				sometree,
				___(irbut.OP_AT, ª(i)),
			)))
		}
	}

	out(prog.Interp(___(ø, ___(irbut.OP_EQ, ___(
		___(irbut.OP_CONST, ª(321)),
		___(irbut.OP_CONST, ª(321)),
	)))))

	out(prog.Interp(___(ø, ___(irbut.OP_INCR, ___(irbut.OP_CONST, ª(22))))))

}
