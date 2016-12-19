package types

import "fmt"

type SliceBuiltin struct {
	Element Type
	methods []Method
}

func (s SliceBuiltin) IsGeneric() bool {
	if a, ok := s.Element.(Generic); ok {
		return a.IsGeneric()
	} else {
		return false
	}
}

func NewSliceBuiltin(element Type, methods []Method) SliceBuiltin {
	return SliceBuiltin{Element: element, methods: methods}
}

func (s *SliceBuiltin) Extend(methods []Method) {
	s.methods = append(s.methods, methods...)
}

func (s SliceBuiltin) Methods() []Method {
	return s.methods
}

func (s SliceBuiltin) Vars() []GenericVar {
	if a, ok := s.Element.(Generic); ok {
		return a.Vars()
	} else {
		return []GenericVar{}
	}
}

func (s SliceBuiltin) IVars() []Type {
	return []Type{}
}

func (s SliceBuiltin) ToString() string {
	return fmt.Sprintf("[]%s", s.Element.ToString())
}

func (s SliceBuiltin) Accepts(t Type) bool {
	if a, ok := t.(SliceBuiltin); ok {
		return s.Element.Accepts(a.Element)
	} else {
		return false
	}
}
