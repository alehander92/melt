package compiler

import (
	"fmt"

	"gitlab.com/alehander42/melt/types"
)

// Signature node
type Signature struct {
	Args   []types.Type
	Return types.Type

	Info
}

// Function node
type Function struct {
	Label     *Label
	Signature *Signature
	Code      *Code
	Args      []Arg

	Info
}

// Arg node
type Arg struct {
	ID   *Label
	Type types.Type

	Info
}

func (f *Function) ToString(depth int) string {
	return fmt.Sprintf("%sFunction %s", Indent(depth), f.Label.Label)
}

func (*Signature) TypeCheck(*Context) error {
	return nil
}

func (f *Function) TypeCheck(ctx *Context) error {
	c := NewContextIn(ctx)
	unhandled := make(map[string]bool)
	c.Unhandled = &unhandled
	ftype, _ := f.meltType.(types.Function)
	c.ReturnType = ftype.Return
	c.Z = ftype.Error

	baba := []Arg{}
	for _, arg := range f.Args {
		if placeholder, ok := arg.Type.(types.Interface); ok {
			kind, err := ctx.Get(placeholder.Label)
			if err != nil {
				return err
			}

			var next types.Type

			switch actual := kind.(type) {
			case types.Interface:
				if len(actual.GenericVars) != len(placeholder.GenericVars) {
					return fmt.Errorf("Refine %s", placeholder.Label)
				} else {
					(&placeholder).Extend(actual.Methods())
					next = placeholder
				}
			case types.Record:
				if len(actual.GenericVars) != len(placeholder.GenericVars) {
					return fmt.Errorf("Wrong number of %s args", placeholder.Label)
				} else {
					next2 := types.Record{Label: actual.Label, Fields: actual.Fields, GenericVars: actual.GenericVars, InstanceVars: actual.InstanceVars}
					(&next2).ReplaceMethods(actual.Methods())
					for i, let := range placeholder.GenericVars {
						next2.InstanceVars[i] = let
					}
					next = next2
				}
			default:
				return fmt.Errorf("Signature")
			}
			baba = append(baba, Arg{ID: arg.ID, Type: next})
		} else {
			baba = append(baba, arg)
		}
		c.Set(arg.ID.Label, baba[len(baba)-1].Type)
	}
	f.Args = baba

	fArgs := []types.Type{}
	for _, arg := range f.Args {
		fArgs = append(fArgs, arg.Type)
	}

	err := f.Code.TypeCheck(c)
	if err != nil {
		return err
	}

	ftype.Args = fArgs
	f.meltType = ftype
	ctx.Set(f.Label.Label, f.meltType)
	return nil
}

func (*Arg) TypeCheck(*Context) error {
	return nil
}
