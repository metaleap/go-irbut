package main

import (
	"github.com/metaleap/go-irbut"
)

const (
	srcSimple = `
main:
	this.>123
`
)

type noun = irbut.Noun
type atom = irbut.NounAtom
type cell = irbut.NounCell

func ___(l noun, r noun) *cell { return &cell{L: l, R: r} }

func main() {
	_ = irbut.Parse(irbut.SrcPrelude+"\n\n"+srcSimple, "main")

	println(irbut.Interp(___(atom(987), ___(irbut.OP_CASE, ___(irbut.False, ___(atom(123), atom(321)))))).String())
}
