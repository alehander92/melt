package generator

import (
	"go/ast"
	"go/token"
	comp "gitlab.com/alehander42/melt/compiler"
)

func GenerateEscalate(e *comp.Escalate, ctx *comp.Context) (ast.Stmt, error) {
	  // return &ast.ReturnStmt{Results: []ast.Expr{&ast.Ident{Name: "nil"}}}, nil

		return &ast.IfStmt{
			Cond: &ast.BinaryExpr{
				X: &ast.Ident{Name: "err"},
				Y: &ast.Ident{Name: "nil"},
				Op: token.NEQ},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ReturnStmt{
						Results: []ast.Expr{
							&ast.Ident{Name: "nil"},
							&ast.Ident{Name: "err"}}}}}}, nil
}
