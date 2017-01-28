package compiler

import (
	"fmt"

	"gitlab.com/alehander42/melt/types"
)

func ReplaceGenericVars(a types.Type, genericMap GenericMap) types.Type {
	i := 0
	if other, ok := a.(types.Function); ok {
		if other.Error == types.Maybe {
			var e types.ErrorFunction
			e = types.Correct
			args := []types.Type{}
			for _, arg := range other.Args {
				args = append(args, replaceInternalGenericVars(arg, &i, genericMap))
				switch argType := arg.(type) {
				default:
					{
					}
				case types.Function:
					{
						switch argNewType := args[len(args)-1].(type) {
						default:
							{
							}
						case types.Function:
							{
								if argType.Error == types.Maybe && argNewType.Error == types.Fail {
									e = types.Fail
									break
								}
							}
						}
					}
				}
			}
			ret := replaceInternalGenericVars(other.Return, &i, genericMap)
			instance := []types.Type{}
			for _, v := range other.GenericVars {
				instance = append(instance, genericMap.Types[v.Label])
			}
			return types.Function{
				Args:         args,
				Return:       ret,
				Error:        e,
				GenericVars:  other.GenericVars,
				InstanceVars: instance}
		}
	}
	return replaceInternalGenericVars(a, &i, genericMap)
}

func replaceInternalGenericVars(t types.Type, errors *int, genericMap GenericMap) types.Type {
	switch other := t.(type) {
	case types.Basic:
		s, ok := genericMap.Types[other.Label]
		if ok {
			return s
		}
		return t
	case types.Function:
		e := other.Error
		if other.Error == types.Maybe {
			fmt.Printf("%s %d\n", other.ToString(), *errors)
			e = genericMap.Errors[*errors]
			*errors += 1
		}
		args := []types.Type{}
		for _, arg := range other.Args {
			args = append(args, replaceInternalGenericVars(arg, errors, genericMap))
		}
		ret := replaceInternalGenericVars(other.Return, errors, genericMap)
		instance := []types.Type{}
		for _, v := range other.GenericVars {
			instance = append(instance, genericMap.Types[v.Label])
		}
		return types.Function{
			Args:         args,
			Return:       ret,
			Error:        e,
			GenericVars:  other.GenericVars,
			InstanceVars: instance}
	case types.Pointer:
		object := replaceInternalGenericVars(other.Object, errors, genericMap)
		return types.Pointer{Object: object}
	case types.Record:
		fields := make(map[string]types.Type)
		for field, t := range other.Fields {
			fields[field] = replaceInternalGenericVars(t, errors, genericMap)
		}
		methods := []types.Method{}
		for _, method := range other.Methods() {
			function, _ := replaceInternalGenericVars(method.Function, errors, genericMap).(types.Function)
			methods = append(methods, types.Method{Label: method.Label, Function: function})
		}
		instance := []types.Type{}
		for _, v := range other.GenericVars {
			instance = append(instance, genericMap.Types[v.Label])
		}
		fmt.Println("ok\n")
		r := types.Record{
			Label:        other.Label,
			Fields:       fields,
			GenericVars:  other.GenericVars,
			InstanceVars: instance}
		r.ReplaceMethods(methods)
		return r
	case types.Interface:
		methods := []types.Method{}
		for _, method := range other.Methods() {
			function, _ := replaceInternalGenericVars(method.Function, errors, genericMap).(types.Function)
			methods = append(methods, types.Method{Label: method.Label, Function: function})
		}
		instance := []types.Type{}
		for _, v := range other.GenericVars {
			instance = append(instance, genericMap.Types[v.Label])
		}
		i := types.Interface{
			Label:        other.Label,
			GenericVars:  other.GenericVars,
			InstanceVars: instance}
		i.Extend(methods)
		return i

	case types.SliceBuiltin:
		methods := []types.Method{}
		for _, method := range other.Methods() {
			function, _ := replaceInternalGenericVars(method.Function, errors, genericMap).(types.Function)
			methods = append(methods, types.Method{Label: method.Label, Function: function})
		}
		element := replaceInternalGenericVars(other.Element, errors, genericMap)
		return types.NewSliceBuiltin(element, methods)

	case types.GenericVar:
		s, ok := genericMap.Types[other.Label]
		if ok {
			// fmt.Printf("  %s\n", s)
			return s
		} else {
			return other
		}
	default:
		return other
	}
}
