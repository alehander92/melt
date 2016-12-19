package generator

import (
	"go/ast"

	comp "gitlab.com/alehander42/melt/compiler"
)

func GenerateModule(m comp.Module, ctx *comp.Context) (*ast.File, error) {
	children := []ast.Decl{}

	module := &ast.File{
		Name:  &ast.Ident{Name: m.Package},
		Decls: children,
		Scope: &ast.Scope{
			Objects: map[string]*ast.Object{}},
		Unresolved: []*ast.Ident{}}

	return module, nil
}
