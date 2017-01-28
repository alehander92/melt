package compiler

import (
	"errors"
	"fmt"

	"gitlab.com/alehander42/melt/types"
)

type TypeMap map[string]types.Type

type GenericMap struct {
	Errors []types.ErrorFunction
	Types  TypeMap
}

func NewGenericMap() GenericMap {
	return GenericMap{Errors: make([]types.ErrorFunction, 0), Types: make(TypeMap)}
}

// Instantiation map
// Have to be generated
type Instantiation struct {
	Functions  map[string][]GenericMap
	Interfaces map[string][]GenericMap
	Records    map[string][]GenericMap
}

type Context struct {
	Values         TypeMap
	Parent         *Context
	Instantiations *Instantiation
	Root           *Context
	IsGeneric      bool
	Dependencies   map[string]map[string][]GenericMap
	Label          string
	Unhandled      *map[string]bool
	ReturnType     types.Type
	Z              types.ErrorFunction
}

func NewContext() Context {
	unhandled := make(map[string]bool)
	return Context{
		Values:         make(TypeMap),
		Parent:         nil,
		Root:           nil,
		Label:          "",
		Instantiations: &Instantiation{Functions: make(map[string][]GenericMap), Records: make(map[string][]GenericMap), Interfaces: make(map[string][]GenericMap)},
		Dependencies:   make(map[string]map[string][]GenericMap),
		Z:              types.Correct,
		Unhandled:      &unhandled,
		IsGeneric:      false}
}

func NewContextIn(parent *Context) *Context {
	var root *Context
	if parent.Root != nil {
		root = parent.Root
	} else {
		root = parent
	}

	return &Context{
		Values:    make(TypeMap),
		Parent:    parent,
		Root:      root,
		Label:     parent.Label,
		Unhandled: parent.Unhandled,
		IsGeneric: parent.IsGeneric}
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
