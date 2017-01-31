package generator

import (
	// "errors"
	"go/ast"

	comp "gitlab.com/alehander42/melt/compiler"
	// "gitlab.com/alehander42/melt/types"
)

func GenerateCall(c *comp.Call, ctx *comp.Context) (ast.Expr, error) {
  f := &ast.Ident{Name: c.Function.Label}
  expressions := []ast.Expr{}
  for _, arg := range c.Args {
      expression, err := GenerateExpr(arg, ctx)
      if err != nil {
          return nil, err
      }

      expressions = append(expressions, expression)
  }

	return &ast.CallExpr{
		Args: expressions,
		Fun: f}, nil
}
