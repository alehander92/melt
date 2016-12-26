package generator

import (
	"go/ast"
	"go/token"

	comp "gitlab.com/alehander42/melt/compiler"
)

func GenerateRecord(r comp.Record, ctx *comp.Context) (*ast.GenDecl, []*ast.Object, error) {
	var fields []*ast.Field
	for _, field := range r.Fields {
		t, err := GenerateType(field.MeltType(), ctx)
		if err != nil {
			return nil, []*ast.Object{}, err
		}

		fields = append(fields,
			&ast.Field{
				Names: []*ast.Ident{ToIdent(field.Label.Label)},
				Type:  t})
	}

	t := &ast.TypeSpec{
		Name: ToIdent(r.Label.Label),
		Type: &ast.StructType{
			Fields: &ast.FieldList{
				List: fields}}}

	obj := &ast.Object{Kind: ast.Typ, Name: r.Label.Label, Decl: t}
	t.Name.Obj = obj

	return &ast.GenDecl{Tok: token.TYPE, Specs: []ast.Spec{t}}, []*ast.Object{obj}, nil
}
