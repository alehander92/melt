package generator

import (
	"go/ast"

	comp "gitlab.com/alehander42/melt/compiler"
)

func GenerateMake(m *comp.Make, ctx *comp.Context) (ast.Expr, error) {
  c, err := GenerateExpr(m.Args[0], ctx)
	if err != nil {
		return nil, err
	}

	t, err := GenerateType(m.Type, ctx)
	if err != nil {
		return nil, err
	}

	return &ast.CallExpr{
		Args: []ast.Expr{t, c},
		Fun: &ast.Ident{Name: "make"}}, nil
}
