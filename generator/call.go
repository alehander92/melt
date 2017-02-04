package generator

import (
	"go/ast"

	comp "gitlab.com/alehander42/melt/compiler"
)

func GenerateCall(c *comp.Call, ctx *comp.Context) (ast.Expr, error) {
  f, err := GenerateLabel(c.Function, ctx)
	if err != nil {
		return nil, err
	}

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
