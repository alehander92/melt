package types

import "fmt"

type Pointer struct {
	Sofia  bool
	Object Type
}

func (self Pointer) ToString() string {
	return fmt.Sprintf("*%s", self.Object.ToString())
}

func (self Pointer) Accepts(t Type) bool {
	switch other := t.(type) {
	case Pointer:
		return self.Object.Accepts(other.Object)
	default:
		return false
	}
}
