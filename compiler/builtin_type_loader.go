package compiler

import (
	"gitlab.com/alehander42/melt/compiler/types"
)

func (env Context) LoadBuiltinTypes() {
	env.Set("map", types.MapBuiltin{})
	intType := types.Basic{Label: "int"}
	stringType := types.Basic{Label: "string"}
	env.Set("Slice", types.NewSliceBuiltin(
		types.GenericVar{Label: "T"},
		[]types.Method{
			{
				Label: "Begin",
				Function: types.Function{
					Return:       types.Pointer{Object: types.GenericVar{Label: "T"}},
					Args:         []types.Type{},
					Error:        types.Correct,
					GenericVars:  []types.GenericVar{},
					InstanceVars: []types.Type{}}},
			{
				Label: "Next",
				Function: types.Function{
					Return:       types.Pointer{Object: types.GenericVar{Label: "T"}},
					Args:         []types.Type{},
					Error:        types.Correct,
					GenericVars:  []types.GenericVar{},
					InstanceVars: []types.Type{}}},
			{
				Label: "Length",
				Function: types.Function{
					Return:       intType,
					Args:         []types.Type{},
					Error:        types.Correct,
					GenericVars:  []types.GenericVar{},
					InstanceVars: []types.Type{}}}}))
	env.Set("int", intType)
	env.Set("string", stringType)
	env.Set("float", types.Basic{Label: "float"})
	env.Set("bool", types.Basic{Label: "bool"})
	env.Set("nil", types.Nil{})
	env.Set("int8", types.Basic{Label: "int8"})
	env.Set("int16", types.Basic{Label: "int16"})
	env.Set("int32", types.Basic{Label: "int32"})
	env.Set("int64", types.Basic{Label: "int64"})
	env.Set("byte", types.Basic{Label: "byte"})
	env.Set("len", types.Function{Return: intType, Error: types.Correct})
	env.Set("print", types.Function{Return: stringType, Error: types.Correct})
}
