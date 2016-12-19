package types

import (
	"fmt"
	"strings"
)

type Record struct {
	Label        string
	methods      []Method
	Fields       map[string]Type
	GenericVars  []GenericVar
	InstanceVars []Type
}

func (r Record) Methods() []Method {
	return r.methods
}

func (r *Record) ReplaceMethods(methods []Method) {
	copy(r.methods, methods)
}

func (r Record) ToString() string {
	var elements []string
	for field, kind := range r.Fields {
		elements = append(elements, field+"."+kind.ToString())
	}
	fields := strings.Join(elements, ",")
	return fmt.Sprintf("#%s{%s}", r.Label, fields)
}

func (r Record) Accepts(t Type) bool {
	switch other := t.(type) {
	case Basic:
		return false
	case Record:
		return r.Label == other.Label
	default:
		return false
	}
}

func (r Record) IsGeneric() bool {
	return len(r.GenericVars) > 0
}

func (r Record) Vars() []GenericVar {
	return r.GenericVars
}

func (r Record) IVars() []Type {
	return r.InstanceVars
}
