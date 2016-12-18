package types

import "fmt"

type Generic interface {
	IsGeneric() bool
	Vars() []GenericVar
	IVars() []Type
	ToString() string
}

type GenericVar struct {
	Label  string
	Actual Type
}

func (self GenericVar) IsGeneric() bool {
	return true
}

func (self GenericVar) ToString() string {
	return fmt.Sprintf("@%s", self.Label)
}

func (self GenericVar) Accepts(t Type) bool {
	switch a := t.(type) {
	case Basic:
		self.Actual = a
		return true
	case GenericVar:
		if self.Label == a.Label {
			self.Actual = a.Actual
			return true
		} else {
			return false
		}
	default:
		i, ok := t.(Generic)
		if ok {
			if !i.IsGeneric() {
				self.Actual = a
				return true
			} else {
				return false
			}
		} else {
			self.Actual = a
			return true
		}
	}
}

func (self GenericVar) Vars() []GenericVar {
	return []GenericVar{self}
}

func (self GenericVar) IVars() []Type {
	return []Type{self.Actual}
}
