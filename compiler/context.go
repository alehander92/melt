package compiler

import (
	"errors"
	"fmt"
	"gitlab.com/alehander42/melt/compiler/types"
)

// Instantiation map
// Have to be generated
type Instantiation struct {
	Functions  map[string][]Function
	Interfaces map[string][]Interface
	Records    map[string][]Record
}

type Context struct {
	Values         map[string]types.Type
	Parent         *Context
	Instantiations *Instantiation
	Root           *Context
	Unhandled      *map[string]bool
	ReturnType     types.Type
	Z              types.ErrorFunction
}

func NewContext() Context {
	unhandled := make(map[string]bool)
	return Context{Values: make(map[string]types.Type), Parent: nil, Root: nil, Instantiations: &Instantiation{}, Z: types.Correct, Unhandled: &unhandled}
}

func NewContextIn(parent *Context) *Context {
	return &Context{Values: make(map[string]types.Type), Parent: parent, Root: parent.Root, Unhandled: parent.Unhandled}
}

func (t *Context) Set(label string, value types.Type) {
	t.Values[label] = value
}

func (t *Context) Get(label string) (types.Type, error) {
	current := t
	for {
		value, ok := current.Values[label]
		if ok {
			return value, nil
		}
		if current.Parent == nil {
			break
		}
		current = current.Parent
	}
	return types.Empty{}, errors.New(fmt.Sprintf("Undefined %s", label))
}

func (t *Context) Contains(label string) bool {
	_, ok := t.Values[label]
	return ok
}
