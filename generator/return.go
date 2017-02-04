package generator

import (
	"go/ast"

	comp "gitlab.com/alehander42/melt/compiler"
)

func GenerateReturn(z *comp.Return, ctx *comp.Context) (ast.Stmt, error) {
	value, err := GenerateExpr(*z.Value, ctx)
	if err != nil {
		  return &ast.ReturnStmt{}, err
	}

  return &ast.ReturnStmt{Results: []ast.Expr{value}}, nil
}
