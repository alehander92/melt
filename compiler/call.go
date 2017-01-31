package compiler

import (
	"errors"
	"fmt"

	"gitlab.com/alehander42/melt/types"
)

type MethodCall struct {
	Receiver *Ast
	Method   *Label
	Args     []Ast

	Info
}

func (m *MethodCall) TypeCheck(ctx *Context) error {
	err := (*m.Receiver).TypeCheck(ctx)
	if err != nil {
		return err
	}

	for _, arg := range m.Args {
		err = arg.TypeCheck(ctx)
		if err != nil {
			return err
		}
	}

	if objectType, ok := (*m.Receiver).MeltType().(types.Duck); ok {
		kind, ok := types.Accepts(objectType, m.Method.Label)
		if !ok {
			return errors.New("method doesn't respond")
		}

		actual, _, err := CallCheck(m.Method.Label, kind.Function, m.Args, &objectType, ctx)
		if err != nil {
			return err
		}

		m.ZType = actual
	} else {
		return errors.New("doesn't have method")
	}
	return nil
}

// Call node
type Call struct {
	Function *Label
	Args     []Ast

	Info
}

func (c *Call) TypeCheck(ctx *Context) error {
	err := c.Function.TypeCheck(ctx)
	if err != nil {
		return err
	}

	for _, arg := range c.Args {
		err = arg.TypeCheck(ctx)
		if err != nil {
			return err
		}
	}

	if function, ok := c.Function.MeltType().(types.Function); ok {
		actual, genericMap, err := CallCheck(c.Function.Label, function, c.Args, nil, ctx)
		if err != nil {
			return err
		}
		(*ctx.Unhandled)[c.Function.Label] = true
		c.ZType = actual

		if len(function.InstanceVars) > 0 {
			if !ctx.IsGeneric {
				fmt.Printf("J %s %d\n", c.Function.Label, len(function.InstanceVars))

				functions, ok := ctx.Root.Instantiations.Functions[c.Function.Label]
				if !ok {
					functions = []GenericMap{}
				}
				ctx.Root.Instantiations.Functions[c.Function.Label] = append(functions,
					genericMap)
			} else {
				fmt.Printf("K %s %d\n", c.Function.Label, len(function.InstanceVars))

				d, ok := ctx.Root.Dependencies[ctx.Label][c.Function.Label]
				if !ok {
					d = []GenericMap{}
				}
				ctx.Root.Dependencies[ctx.Label][c.Function.Label] = d
			}
		}

	} else {
		return errors.New("not a function type")
	}

	return nil
}

func CallCheck(label string, function types.Function, args []Ast, receiver *types.Duck, ctx *Context) (types.Type, GenericMap, error) {

	if label == "len" && receiver == nil {
		return LenCheck(function, args, ctx)
	} else if label == "print" && receiver == nil {
		p := function.Return
		return p, GenericMap{}, nil
	}

	if len(function.Args) != len(args) {
		return types.Empty{},
			GenericMap{},
			fmt.Errorf("Expected different args %s:\n    received %d\n    wanted %d", label, len(args), len(function.Args))
	}

	error := types.Correct
	if label[len(label)-1] == '!' {
		error = types.Fail
	} else if label[len(label)-1] == '?' {
		error = types.Maybe
	}

	if function.Error == types.Correct && error != types.Correct ||
		function.Error == types.Fail && error == types.Correct {
		return types.Empty{}, GenericMap{}, fmt.Errorf("Error %s: received %s, wanted %s", label, types.Alexander(error), types.Alexander(function.Error))
	}

	if len(function.GenericVars) > 0 {
		genericMap := NewGenericMap()
		for _, r := range function.GenericVars {
			genericMap.Types[r.Label] = types.Empty{}
		}
		for i, arg := range args {
			fArg := function.Args[i]
			err := Match(&genericMap, arg.MeltType(), fArg, ctx)
			if err != nil {
				return types.Empty{}, GenericMap{}, err
			}
		}

		for id := range genericMap.Types {
			_, ok := genericMap.Types[id].(types.Empty)
			if ok {
				return types.Empty{}, GenericMap{}, fmt.Errorf("Error %s: %s not actualized", label, id)
			}
		}

		returnType := ReplaceGenericVars(function.Return, genericMap)

		return returnType, genericMap, nil
	} else {
		for i, arg := range args {
			fArg := function.Args[i]
			if !fArg.Accepts(arg.MeltType()) {
				return types.Empty{}, GenericMap{}, fmt.Errorf("Bad call:\n    received %s\n    wanted %s", arg.MeltType().ToString(), fArg.ToString())
			}
		}
		return function.Return, GenericMap{}, nil
	}
}

func LenCheck(function types.Function, args []Ast, ctx *Context) (types.Type, GenericMap, error) {
	if len(args) != 1 {
		return types.Empty{}, GenericMap{}, errors.New("Len takes one arg")
	} else {
		switch a := args[0].MeltType().(type) {
		case types.SliceBuiltin:
			i := function.Return
			return i, GenericMap{}, nil
		case types.Duck:
			length, ok := types.Accepts(a, "Length")
			j, k := a.(types.Interface)
			fmt.Printf("  %s %s\n", j.Methods(), k)
			if ok {
				if len(length.Function.Args) == 0 && length.Function.Error == types.Correct {
					m, ok := length.Function.Return.(types.Basic)
					if ok && m.Label == "int" {
						return m, GenericMap{}, nil
					}
				}
			}
			return types.Empty{}, GenericMap{}, errors.New("Length() int")
		default:
			return types.Empty{}, GenericMap{}, errors.New("Slice or Length")
		}
	}
}

func Match(genericMap *GenericMap, callArg types.Type, arg types.Type, ctx *Context) error {
	// fmt.Printf("%s\n", arg.ToString())
	c, ok := callArg.(types.SliceBuiltin)
	if ok {
		s, _ := ctx.Get("Slice")
		s2, _ := s.(types.SliceBuiltin)
		c.Extend(s2.Methods())
		m := NewGenericMap()
		m.Types["T"] = c.Element
		callArg = ReplaceGenericVars(c, m)
		t, _ := callArg.(types.SliceBuiltin)
		fmt.Printf("  %s\n", t.Methods()[0].Function.ToString())
	}

	switch other := arg.(type) {
	case types.Basic:
		o, ok := (*genericMap).Types[other.Label]
		if ok {
			_, ok := o.(types.Empty)
			if ok {
				(*genericMap).Types[other.Label] = callArg
				return nil
			} else {
				if !o.Accepts(callArg) {
					return fmt.Errorf("received %s, wanted %s", callArg.ToString(), o.ToString())
				}
			}
		} else {
			if !arg.Accepts(callArg) {
				return fmt.Errorf("received %s, wanted %s", callArg.ToString(), arg.ToString())
			} else {
				return nil
			}
		}

	case types.GenericVar:
		o, ok := (*genericMap).Types[other.Label]
		if ok {
			_, ok := o.(types.Empty)
			if ok {
				(*genericMap).Types[other.Label] = callArg
				return nil
			} else {
				if !o.Accepts(callArg) {
					return fmt.Errorf("%s is %s, can't be %s", other.Label, o.ToString(), callArg.ToString())
				} else {
					return nil
				}
			}
		} else {
			return fmt.Errorf("unknown %s", other.Label)
		}

	case types.Record:
		o, ok := callArg.(types.Record)
		if !ok {
			return fmt.Errorf("%s is not a record", callArg.ToString())
		}
		if other.Label != o.Label && len(other.InstanceVars) != len(o.InstanceVars) {
			return fmt.Errorf("%s is not %s", callArg.ToString(), arg.ToString())
		}
		for i, arg := range o.InstanceVars {
			err := Match(genericMap, arg, other.InstanceVars[i], ctx)
			if err != nil {
				return err
			}
		}

	case types.Function:
		o, ok := callArg.(types.Function)
		if !ok {
			return fmt.Errorf("%s is not a function", callArg.ToString())
		}
		if other.Error == types.Correct && o.Error != types.Correct ||
			other.Error == types.Fail && o.Error != types.Fail {
			return fmt.Errorf("%s fix fail", callArg.ToString())
		}
		if other.Error == types.Maybe {
			if o.Error != types.Maybe {
				genericMap.Errors = append(genericMap.Errors, o.Error)
			} else {
				fmt.Printf("ERRORS %s", callArg.ToString())
				genericMap.Errors = append(genericMap.Errors, types.Maybe)
			}
		}
		if len(other.Args) != len(o.Args) {
			return fmt.Errorf("%s fix arity", callArg.ToString())
		}
		for i, arg := range o.Args {
			err := Match(genericMap, arg, other.Args[i], ctx)
			if err != nil {
				return err
			}
		}
		r := o.Return
		err := Match(genericMap, r, other.Return, ctx)
		if err != nil {
			return err
		}

	case types.Interface:
		duck, ok := callArg.(types.Duck)
		if !ok {
			return fmt.Errorf("%s not a duck", callArg.ToString())
		}

		fmt.Printf("%T\n", other.Methods())
		for _, m := range other.Methods() {
			fmt.Printf("  #%s\n", m.Label)

			value, ok := types.Accepts(duck, m.Label)
			if !ok {
				return errors.New("not valid")
			}

			function := value.Function
			if function.Error != m.Function.Error ||
				len(function.Args) != len(m.Function.Args) {
				return fmt.Errorf("%s is different", callArg.ToString())
			}

			for i, arg := range m.Function.Args {
				err := Match(genericMap, function.Args[i], arg, ctx)
				if err != nil {
					return err
				}
			}

			err := Match(genericMap, function.Return, m.Function.Return, ctx)
			if err != nil {
				return err
			}
		}
		return nil
	case types.Pointer:
		t, ok := callArg.(types.Pointer)
		if !ok {
			return fmt.Errorf("%s is not a pointer", callArg.ToString())
		}
		return Match(genericMap, t.Object, other.Object, ctx)
	default:
		if !arg.Accepts(callArg) {
			return fmt.Errorf("received %s, wanted %s", callArg.ToString(), ReplaceGenericVars(arg, *genericMap).ToString())
		} else {
			return nil
		}
	}
	return nil
}
