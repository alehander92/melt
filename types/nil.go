package types

import "fmt"

type Nil struct {
}

func (self Nil) ToString() string {
	return "nil"
}

func (self Nil) Accepts(t Type) bool {
	a, ok := t.(Nil)
	fmt.Printf("%s", a)
	return ok
}
