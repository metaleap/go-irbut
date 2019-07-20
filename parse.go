package irbut

const SrcPrelude = `

feed f1 f2 :=
	this.f1.f2

keep f1 f2 :=
	[this.f1 this].f2

call addr f :=
	this.f.![@1 @addr]

case boolish ifTrue ifFalse :=


	`

func Parse(src string) (Noun, error) {
	return nil, nil
}
