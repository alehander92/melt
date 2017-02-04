package generator

import (
	// "errors"
	"go/ast"

	comp "gitlab.com/alehander42/melt/compiler"
	// "gitlab.com/alehander42/melt/types"
)

func GenerateLabel(l *comp.Label, ctx *comp.Context) (ast.Expr, error) {
  var label *ast.Ident
  if l.Label[len(l.Label) - 1] == '?' || l.Label[len(l.Label) - 1] == '!' {
    label = &ast.Ident{Name: l.Label[0:len(l.Label) - 1]}
  } else {
    label = &ast.Ident{Name: l.Label}
  }
  return label, nil
}
