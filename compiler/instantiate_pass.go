package compiler

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/alehander42/deepcopy"
	"gitlab.com/alehander42/melt/types"
)

func Instantiate(m *Module, ctx *Context) error {
	fmt.Printf("DEPENDENCIES\n%s\n", ctx.Dependencies)
	fmt.Printf("INSTANTIATIONS\n%s\n", ctx.Instantiations.Functions["Map"][0].Errors)
	x := deepcopy.Copy(m)
	fmt.Printf("MODULE %s\n", x)
	expanded := make(map[string]map[string]Function)
	functions := make(map[string]Function)

	for _, f := range m.Functions {
		g, ok := ctx.Dependencies[f.Label.Label]
		if ok {
			err := ExpandDependencies(&g, f, functions, ctx)
			if err != nil {
				return err
			}
		}
	}

	for _, f := range m.Functions {
		functions[f.Label.Label] = *f
		expanded[f.Label.Label] = make(map[string]Function)
	}

	for _, f := range m.Functions {
		i, ok := ctx.Instantiations.Functions[f.Label.Label]
		g, ok2 := ctx.Dependencies[f.Label.Label]
		if ok2 {
		}
		if ok {
			for _, in := range i {
				label := FunctionName(*f, in)
				sex, ok := expanded[f.Label.Label]

				if !ok {
					expanded[f.Label.Label] = make(map[string]Function)
					sex, _ = expanded[f.Label.Label]
				}
				_, ok = sex[label]
				if ok {
					continue
				}

				exp, err := ExpandInstance(*f, len(sex), in)
				if err != nil {
					return err
				}

				expanded[f.Label.Label][label] = exp
				for l, dep := range g {
					functionDep, ok := functions[l]
					if !ok {
						return fmt.Errorf("%s is missing", l)
					}
					for _, d := range dep {
						label := FunctionName(functionDep, d)
						sex, ok := expanded[functionDep.Label.Label]
						if !ok {
							sex = make(map[string]Function)
							expanded[functionDep.Label.Label] = sex
						}
						_, ok = sex[label]
						if ok {
							continue
						}
						exp, err := ExpandInstance(functionDep, len(sex), d)
						if err != nil {
							return err
						}
						expanded[functionDep.Label.Label][label] = exp
					}
				}
			}
		}
	}
	funs := []*Function{}
	for _, v := range expanded {
		for _, x := range v {
			funs = append(funs, &x)
		}
	}
	m.Functions = funs

	return nil
}

func ExpandDependencies(dependencies *map[string][]GenericMap, function *Function, functions map[string]Function, ctx *Context) error {
	f, ok := function.MeltType().(types.Function)
	if !ok {
		return fmt.Errorf("err")
	}
	fmt.Printf("  %s %s\n", f, dependencies)
	return nil
}

func ExpandInstance(function Function, index int, genericMap GenericMap) (Function, error) {
	// fmt.Printf("%s @\n", genericMap)
	fun := Walk(function, true, func(node Ast) {
		t := ReplaceGenericVars(node.MeltType(), genericMap)
		node.ChangeMeltType(t)
		fmt.Printf("before:%s after:%s %s\n", node.MeltType().ToString(), t.ToString(), genericMap)
	})
	fun.Label.Label = fmt.Sprintf("%s%d", function.Label.Label, index)
	f, ok := fun.MeltType().(types.Function)
	if !ok {
		return Function{}, fmt.Errorf("Sick function")
	}
	fmt.Printf("Expand %s\n", f.ToString())

	if f.Error == types.Maybe {
		e := types.Correct
		for _, arg := range function.Args {
			if b, ok := arg.Type.(types.Function); ok {
				if b.Error == types.Fail {
					e = types.Fail
					break
				}
			}
		}
		f.Error = e
	}
	fun.meltType = f
	// fmt.Printf("type %s\n", fun.MeltType().ToString())
	return fun, nil
}

func FunctionName(function Function, genericMap GenericMap) string {
	args := []string{}
	for arg, kind := range genericMap.Types {
		args = append(args, fmt.Sprintf("[%s %s]", arg, kind.ToString()))
	}
	for _, e := range genericMap.Errors {
		args = append(args, fmt.Sprintf("{%s}", e))
	}
	return strings.Join(args, "")
}

func Clone(node Ast) Ast {
	z := CloneValue(node)
	y, _ := z.Interface().(Ast)
	return y
}

func CloneValue(node Ast) reflect.Value {
	r := reflect.ValueOf(&node)
	s := r.Elem().Elem().Elem()
	typeOfNode := s.Type()

	z := reflect.New(typeOfNode)
	// fmt.Printf("%s\n", s.Type())
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		switch f.Kind() {
		case reflect.Ptr:
			g := f.Elem()
			if g.Kind() == reflect.Struct {
				x, ok := f.Interface().(Ast)
				// fmt.Printf("%s %s\n", x, f.Kind())
				if !ok {
					z.Elem().Field(i).Set(f)
				} else {
					// p := reflect.New(f.Type())
					// p.Elem().Set(CloneValue(x))
					z.Elem().Field(i).Set(CloneValue(x))
				}
			} else {
				z.Elem().Field(i).Set(f)
			}
		case reflect.Struct:
			x, ok := f.Interface().(Ast)
			// fmt.Printf("%s %s\n", x, f)
			if !ok {
				z.Elem().Field(i).Set(f)
			} else {
				z.Field(i).Set(CloneValue(x))
			}
		case reflect.Slice:
			fmt.Printf("SLICE\n")
			fmt.Printf("%d %s\n", f.Len(), f.Type())
			slice := reflect.New(f.Type())
			for j := 0; j < f.Len(); j++ {
				element := f.Index(j)
				x, ok := element.Interface().(Ast)
				if !ok {
					reflect.Append(slice.Elem(), element)
				} else {
					reflect.Append(slice.Elem(), CloneValue(x))
				}
			}
			z.Elem().Field(i).Set(slice.Elem())
		case reflect.Map:
			m := reflect.New(f.Type())
			keys := f.MapKeys()
			for _, key := range keys {
				value := f.MapIndex(key)
				x, ok := value.Interface().(Ast)
				if !ok {
					m.SetMapIndex(key, value)
				} else {
					m.SetMapIndex(key, CloneValue(x))
				}
			}
			z.Elem().Field(i).Set(m)
		default:
			z.Elem().Field(i).Set(f)
		}
		// typeOfNode.Field(i)
	}

	return z
}

func Walk(function Function, clone bool, handler func(Ast)) Function {
	f := &function
	var ok bool
	if clone {
		f0 := Pre(&function)
		f, ok = f0.(*Function)
		if !ok {
			return Function{}
		}
	}
	handler(f)
	return *f
}

func Pre(a Ast) Ast {
	b := deepcopy.Copy(a)
	c, ok := b.(Ast)
	if !ok {
		fmt.Printf("Ops\n")
		return &Module{}
	}
	c.ChangeMeltType(a.MeltType())
	fmt.Printf("BEFORE  %s\n  %s\n", a.MeltType().ToString(), c.MeltType().ToString())

	return c
}
