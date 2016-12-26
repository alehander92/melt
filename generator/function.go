package generator

import (
	"errors"
	"fmt"
	"go/ast"

	comp "gitlab.com/alehander42/melt/compiler"
	"gitlab.com/alehander42/melt/types"
)

func GenerateFunction(f comp.Function, ctx *comp.Context) (*ast.FuncDecl, []*ast.Object, error) {
	block, err := GenerateCode(f.Code, ctx)
	if err != nil {
		return nil, []*ast.Object{}, err
	}

	fields := []*ast.Field{}
	results := []*ast.Field{}
	for _, arg := range f.Args {
		t, err := GenerateType(arg.Type, ctx)
		if err != nil {
			return nil, []*ast.Object{}, err
		}

		fields = append(fields,
			&ast.Field{
				Names: []*ast.Ident{ToIdent(arg.ID.Label)},
				Type:  t})
	}

	m, ok := f.MeltType().(types.Function)
	if !ok {
		return nil, []*ast.Object{}, errors.New("Invalid")
	}

	returnType, err := GenerateType(m.Return, ctx)
	if err != nil {
		return nil, []*ast.Object{}, err
	}

	results = append(results, &ast.Field{Type: returnType})

	if m.Error == types.Correct {
		fmt.Println("yes")
	} else if m.Error == types.Maybe {
		return nil, []*ast.Object{}, errors.New("? impossible")
	} else {
		results = append(results, &ast.Field{Type: ToIdent("error")})
	}

	f2 := &ast.FuncDecl{
		Name: ToIdent(f.Label.Label),
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: fields},
			Results: &ast.FieldList{
				List: results}},
		Body: block}

	obj := &ast.Object{Kind: ast.Fun, Name: f.Label.Label, Decl: f2}
	f2.Name.Obj = obj
	return f2, []*ast.Object{obj}, nil
}
