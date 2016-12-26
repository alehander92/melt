package generator

import (
	"go/ast"

	comp "gitlab.com/alehander42/melt/compiler"
)

func GenerateModule(m comp.Module, ctx *comp.Context) (*ast.File, error) {
	children := []ast.Decl{}
	objects := make(map[string]*ast.Object)

	// for _, child := range m.Interfaces {
	// 	a, err := GenerateInterface(child, ctx)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	children = append(children, a)
	// }

	for _, child := range m.Records {
		record, objs, err := GenerateRecord(child, ctx)
		if err != nil {
			return nil, err
		}

		children = append(children, record)
		for _, obj := range objs {
			objects[obj.Name] = obj
		}
	}

	for _, child := range m.Functions {
		function, objs, err := GenerateFunction(child, ctx)
		if err != nil {
			return nil, err
		}

		children = append(children, function)
		for _, obj := range objs {
			objects[obj.Name] = obj
		}
	}

	module := &ast.File{
		Name:  &ast.Ident{Name: m.Package},
		Decls: children,
		Scope: &ast.Scope{
			Objects: objects},
		Unresolved: []*ast.Ident{}}

	return module, nil
}
