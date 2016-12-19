package types

import "fmt"

type Type interface {
	Accepts(Type) bool
	ToString() string
}

func ReplaceGenericVars(t Type, genericMap map[string]Type) Type {
	fmt.Printf("%s %s\n", t.ToString(), genericMap)
	switch other := t.(type) {
	case Basic:
		return t
	case Function:
		args := []Type{}
		for _, arg := range other.Args {
			args = append(args, ReplaceGenericVars(arg, genericMap))
		}
		ret := ReplaceGenericVars(other.Return, genericMap)
		instance := []Type{}
		for _, v := range other.GenericVars {
			instance = append(instance, genericMap[v.Label])
		}
		return Function{
			Args:         args,
			Return:       ret,
			Error:        other.Error,
			GenericVars:  other.GenericVars,
			InstanceVars: instance}
	case Pointer:
		object := ReplaceGenericVars(other.Object, genericMap)
		return Pointer{Object: object}
	case Record:
		fields := make(map[string]Type)
		for field, t := range other.Fields {
			fields[field] = ReplaceGenericVars(t, genericMap)
		}
		methods := []Method{}
		for _, method := range other.methods {
			function, _ := ReplaceGenericVars(method.Function, genericMap).(Function)
			methods = append(methods, Method{Label: method.Label, Function: function})
		}
		instance := []Type{}
		for _, v := range other.GenericVars {
			instance = append(instance, genericMap[v.Label])
		}
		fmt.Println("ok\n")
		return Record{
			Label:        other.Label,
			Fields:       fields,
			methods:      methods,
			GenericVars:  other.GenericVars,
			InstanceVars: instance}
	case Interface:
		methods := []Method{}
		for _, method := range other.methods {
			function, _ := ReplaceGenericVars(method.Function, genericMap).(Function)
			methods = append(methods, Method{Label: method.Label, Function: function})
		}
		instance := []Type{}
		for _, v := range other.GenericVars {
			instance = append(instance, genericMap[v.Label])
		}
		return Interface{
			Label:        other.Label,
			methods:      methods,
			GenericVars:  other.GenericVars,
			InstanceVars: instance}
	case SliceBuiltin:
		methods := []Method{}
		for _, method := range other.methods {
			function, _ := ReplaceGenericVars(method.Function, genericMap).(Function)
			methods = append(methods, Method{Label: method.Label, Function: function})
		}
		return SliceBuiltin{
			Element: other.Element,
			methods: methods}
	case GenericVar:
		s, ok := genericMap[other.Label]
		if ok {
			fmt.Printf("  %s\n", s)
			return s
		} else {
			return other
		}
	default:
		return other
	}
}
