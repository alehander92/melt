package generator

import (
	"go/ast"
	"go/token"
	comp "gitlab.com/alehander42/melt/compiler"
)

func GenerateForIn(f *comp.ForIn, ctx *comp.Context) (ast.Stmt, error) {
	  // return &ast.ReturnStmt{Results: []ast.Expr{}}, nil

	  var key *ast.Ident;
	  var value *ast.Ident;
	  if len(f.Index) == 1 {
		    key = &ast.Ident{Name: "_"}
		    value = &ast.Ident{Name: f.Index[0].Label}
	  } else {
			  key = &ast.Ident{Name: f.Index[0].Label}
				value = &ast.Ident{Name: f.Index[1].Label}
		}

		sequence, err := GenerateExpr(*f.Sequence, ctx)
		if err != nil {
			  return nil, err
		}
		var b []ast.Stmt = []ast.Stmt{}
		for _, e := range f.Code.E {
			  node, err := GenerateNode(e, ctx)
				if err != nil {
					  return nil, err
				}
				b = append(b, node)
		}
		return &ast.RangeStmt{
			  Value: value,
				Key: key,
				Tok: token.DEFINE,
				X: sequence,
				Body: &ast.BlockStmt{List: b}}, nil
		// 		// Value: labelInit: , nil
}
