package compiler

import (
	"fmt"
	"gitlab.com/alehander42/melt/types"
)

func Instantiate(m *Module, ctx *Context) error {
	fmt.Printf("DEPENDENCIES\n%s\n", ctx.Dependencies)
	fmt.Printf("INSTANTIATIONS\n%s\n", ctx.Instantiations)
	expanded := make(map[string]Function)
	for _, f := range m.Functions {
		g, ok := ctx.Dependencies[f.Label.Label]
		if ok {
			ExpandDependencies(g, f, m.Functions, ctx)
		}
	}

	for _, f := range m.Functions {
		i, ok := ctx.Instantiations[f.Label.Label]
		g, ok2 := ctx.Dependencies[f.Label.Label]
		d := []map[string][]map[string]types.Type{}
		if ok2 {
			d = g
		}
		if ok {
			for _, inst := range i {
				ExpandInstantation(f, inst)

				for _, dep := range d {
					ExpandInstantation(f, dep)
				}
		}
	}
	functions := []Function{}
	for _, v := range expanded {
		functions = append(functions, v)
	}
	m.Functions = functions

	return nil
}

func ExpandDependencies(dependencies *map[string][]map[string]types.Type, function Function, functions []Function, ctx *Context) {
}
