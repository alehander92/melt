package compiler

import (
	"fmt"

	"gitlab.com/alehander42/melt/types"
)

func (self *Context) CollectTypes(ast Module) error {
	a := []string{}
	types := []types.Type{}
	for _, i := range ast.Interfaces {
		a = append(a, i.Label.Label)
		types = append(types, i.MeltType())
	}
	err := self.collectFrom(a, types, "Interface")
	if err != nil {
		return err
	}

	a = a[:0]
	types = types[:0]

	for _, r := range ast.Records {
		a = append(a, r.Label.Label)
		types = append(types, r.MeltType())
	}
	err = self.collectFrom(a, types, "Record")
	if err != nil {
		return err
	}

	a = a[:0]
	types = types[:0]
	for _, f := range ast.Functions {
		a = append(a, f.Label.Label)
		types = append(types, f.MeltType())
		if len(f.Args) > 1 {
			fmt.Printf("%s\n", f.MeltType().ToString())
			g := x(f.MeltType())
			fmt.Printf("%s\n", g.Methods())
		}
	}
	err = self.collectFrom(a, types, "Function")
	return err
}

func (self *Context) collectFrom(nodes []string, types []types.Type, label string) error {
	for i, node := range nodes {
		if !self.Contains(node) {
			self.Set(node, types[i])
		} else {
			return fmt.Errorf("%s %s can't be redefined", label, node)
		}
	}
	return nil
}

func x(t types.Type) types.Interface {
	g, _ := t.(types.Function)
	g2, _ := g.Args[1].(types.Interface)
	return g2
}
