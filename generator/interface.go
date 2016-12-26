package generator

import (
	"go/ast"

	comp "gitlab.com/alehander42/melt/compiler"
)

func GenerateInterface(i comp.Interface, ctx *comp.Interface) (*ast.GenDecl, []*ast.Object, error) {
	return &ast.GenDecl{}, []*ast.Object{}, nil
}
