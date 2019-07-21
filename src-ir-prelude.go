package irbut

const SrcPrelude = `

id:
	@this

konst val:
	@val

konst4 val1 val2 val3 val4:
	@val2.id

// composition
feed f1 f2:
	this.f1.f2

// extension
keep f1 f2:
	// wot now
	[this.f1 this].f2
	123 : call

// invocation
call addr f:
	this.f.![@1 @addr]

// *[this *[ [ifTrue ifFalse] 0 *[[2 3] 0 *[this 4 4 boolish]]]]
case boolish ifTrue ifFalse:
	this.foo
	foo: [ifTrue ifFalse].@addr
	addr: [2 3].@this.++boolish

	`
