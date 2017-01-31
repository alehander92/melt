package generator

import (
	// "errors"
	"go/ast"

	comp "gitlab.com/alehander42/melt/compiler"
	// "gitlab.com/alehander42/melt/types"
)

func GenerateLabel(l *comp.Label, ctx *comp.Context) (ast.Expr, error) {
  label := &ast.Ident{Name: l.Label}
  return label, nil
}
