package generator

import (
	"fmt"
	"go/ast"

	comp "gitlab.com/alehander42/melt/compiler"
)

func GenerateCode(c *comp.Code, ctx *comp.Context) (*ast.BlockStmt, error) {
	list := []ast.Stmt{}
	for _, code := range c.E {
		fmt.Printf("%s\n", code.MeltType())
		expr, err := GenerateNode(code, ctx)
		if err != nil {
			return nil, err
		}
		if expr == nil {
			continue
		}
		list = append(list, expr)
	}
	return &ast.BlockStmt{
		List: list}, nil
}
