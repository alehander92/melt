package generator

import (
	"errors"
	"fmt"
	"go/ast"

	comp "gitlab.com/alehander42/melt/compiler"
	"gitlab.com/alehander42/melt/types"
)

func GenerateType(t types.Type, ctx *comp.Context) (ast.Expr, error) {
	switch other := t.(type) {
	case types.Basic:
		return &ast.Ident{Name: other.Label}, nil

	case types.SliceBuiltin:
		element, err := GenerateType(other.Element, ctx)
		if err != nil {
			return nil, err
		}

		return &ast.ArrayType{Elt: element}, nil

	case types.Pointer:
		object, err := GenerateType(other.Object, ctx)
		if err != nil {
			return nil, err
		}

		return &ast.StarExpr{X: object}, nil

	case types.Interface:
		l := &ast.Ident{Name: other.Label}
		return l, nil

	case types.Function:
		if other.Error == types.Maybe {
			return nil, fmt.Errorf("%s can't be ?", other.ToString())
		}

		params := []*ast.Field{}
		results := []*ast.Field{}
		for _, arg := range other.Args {
			t, err := GenerateType(arg, ctx)
			if err != nil {
				return nil, err
			}

			params = append(params, &ast.Field{Type: t})
		}

		t, err := GenerateType(other.Return, ctx)
		if err != nil {
			return nil, err
		}

		results = append(results, &ast.Field{Type: t})
		if other.Error == types.Fail {
			results = append(results, &ast.Field{Type: ToIdent("error")})
		}

		return &ast.FuncType{
			Params:  &ast.FieldList{List: params},
			Results: &ast.FieldList{List: results}}, nil

	default:
		fmt.Printf("X:%s\n", t)
		return nil, errors.New("unknown")
	}
}
