package types

import (
	"fmt"
	"strings"
)

//Interface type
type Interface struct {
	Label        string
	methods      []Method
	GenericVars  []GenericVar
	InstanceVars []Type
}

func NewInterface(label string, methods []Method, vars []GenericVar) Interface {
	return Interface{Label: label, methods: methods, GenericVars: vars, InstanceVars: make([]Type, len(vars))}
}

type Duck interface {
	Methods() []Method
	ToString() string
}

func (i *Interface) Extend(methods []Method) {
	i.methods = append(i.methods, methods...)
}

func Accepts(d Duck, label string) (Method, bool) {
	for _, kind := range d.Methods() {
		if kind.Label == label {
			return kind, true
		}
	}
	return Method{}, false
}

func (i Interface) ToString() string {
	g := ""
	genericVars := []string{}
	for _, v := range i.InstanceVars {
		genericVars = append(genericVars, v.ToString())
	}
	g = strings.Join(genericVars, ",")
	if len(i.InstanceVars) > 0 {
		g = fmt.Sprintf("<%s>", g)
	}
	return fmt.Sprintf("%s%s", i.Label, g)
}

func (i Interface) Methods() []Method {
	return i.methods
}

func (i Interface) Accepts(t Type) bool {
	switch other := t.(type) {
	case Basic:
		return len(i.methods) == 0
	case InterfaceOrRecord:
		for _, signature := range i.methods {
			if !other.AcceptsFunction(signature.Function) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func (i Interface) AcceptsFunction(f Function) bool {
	for _, signature := range i.methods {
		if signature.Function.Accepts(f) {
			return true
		}
	}
	return false
}

func (i Interface) IsGeneric() bool {
	return len(i.GenericVars) > 0
}

func (i Interface) Vars() []GenericVar {
	return i.GenericVars
}

func (i Interface) IVars() []Type {
	return i.InstanceVars
}
