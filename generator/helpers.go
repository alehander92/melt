package generator

import (
	"go/ast"

	"gitlab.com/alehander42/melt/types"
)

func ToIdent(label string) *ast.Ident {
	return &ast.Ident{Name: label}
}

func ToType(t types.Type) ast.Expr {
	switch other := t.(type) {
	case types.Basic:
		return &ast.Ident{Name: other.Label}
	}
	return nil
}
