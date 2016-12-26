package generator

import (
	"go/ast"

	comp "gitlab.com/alehander42/melt/compiler"
)

func GenerateCode(c *comp.Code, ctx *comp.Context) (*ast.BlockStmt, error) {
	return &ast.BlockStmt{}, nil
}
