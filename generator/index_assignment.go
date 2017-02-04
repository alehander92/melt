package generator

import (
  "go/ast"
  "go/token"
  comp "gitlab.com/alehander42/melt/compiler"
)

func GenerateIndexAssignment(g *comp.IndexAssignment, ctx *comp.Context) (ast.Stmt, error) {
    c, err := GenerateExpr(*g.Collection, ctx)
    if err != nil {
        return nil, err
    }

    i, err := GenerateExpr(*g.Index, ctx)
    if err != nil {
        return nil, err
    }

    v, err := GenerateExpr(*g.Value, ctx)
    if err != nil {
        return nil, err
    }

    return &ast.AssignStmt{
      Lhs: []ast.Expr{
          &ast.IndexExpr{X: c, Index: i}},
      Rhs: []ast.Expr{
          v},
      Tok: token.ASSIGN}, nil
    return &ast.ReturnStmt{Results: []ast.Expr{}}, nil
}
