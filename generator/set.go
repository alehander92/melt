package generator

import (
	"go/ast"
	"go/token"

	comp "gitlab.com/alehander42/melt/compiler"
)

func GenerateSet(set *comp.Set, ctx *comp.Context) (ast.Stmt, error) {
	value, err := GenerateExpr(*set.Value, ctx)
	if err != nil {
		return nil, err
	}
	a, err := GenerateType((*set.Value).MeltType(), ctx)
	if err != nil {
		return nil, err
	}
	return &ast.DeclStmt{
		Decl: &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{&ast.ValueSpec{
				Names:  []*ast.Ident{{Name: set.Label.Label}},
				Type:   a,
				Values: []ast.Expr{value}}}}}, nil
}
