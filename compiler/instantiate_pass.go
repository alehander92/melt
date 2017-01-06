package compiler

import (
	"fmt"
	"strings"

	"gitlab.com/alehander42/melt/types"
)

func Instantiate(m *Module, ctx *Context) error {
	fmt.Printf("DEPENDENCIES\n%s\n", ctx.Dependencies)
	fmt.Printf("INSTANTIATIONS\n%s\n", ctx.Instantiations)
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
		functions[f.Label.Label] = f
		expanded[f.Label.Label] = make(map[string]Function)
	}

	for _, f := range m.Functions {
		i, ok := ctx.Instantiations.Functions[f.Label.Label]
		g, ok2 := ctx.Dependencies[f.Label.Label]
		if ok2 {
		}
		if ok {
			for _, in := range i {
				label := FunctionName(f, in)
				sex, ok := expanded[f.Label.Label]

				if !ok {
					expanded[f.Label.Label] = make(map[string]Function)
					sex, _ = expanded[f.Label.Label]
				}
				_, ok = sex[label]
				if ok {
					continue
				}

				exp, err := ExpandInstance(f, len(sex), in)
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
	funs := []Function{}
	for _, v := range expanded {
		for _, x := range v {
			funs = append(funs, x)
		}
	}
	m.Functions = funs

	return nil
}

func ExpandDependencies(dependencies *map[string][]TypeMap, function Function, functions map[string]Function, ctx *Context) error {
	f, ok := function.MeltType().(types.Function)
	if !ok {
		return fmt.Errorf("err")
	}
	fmt.Printf("%s %s\n", f, dependencies)
	return nil
}

func ExpandInstance(function Function, index int, genericMap TypeMap) (Function, error) {
	fmt.Printf("%s @\n", genericMap)
	fun := Walk(function, true, func(node Ast) {
		node.ChangeMeltType(types.ReplaceGenericVars(node.MeltType(), genericMap))
	})
	fun.Label.Label = fmt.Sprintf("%s%d", function.Label.Label, index)
	f, ok := function.MeltType().(types.Function)
	if !ok {
		return Function{}, fmt.Errorf("Sick function")
	}
	if f.Error == types.Maybe {
		_, ok := genericMap["Error"]
		if !ok {
			f.Error = types.Correct
		} else {
			f.Error = types.Fail
		}
	}
	fun.meltType = f
	fmt.Printf("type %s\n", fun.MeltType().ToString())
	return fun, nil
}

func FunctionName(function Function, genericMap TypeMap) string {
	args := []string{}
	for arg, kind := range genericMap {
		args = append(args, fmt.Sprintf("[%s %s]", arg, kind.ToString()))
	}
	return strings.Join(args, "")
}

func Clone(node Ast) Ast {
	return node
}

func Walk(function Function, clone bool, handler func(Ast)) Function {
	handler(&function)
	return Function{
		Label:     &Label{Label: function.Label.Label},
		Code:      function.Code,
		Args:      function.Args,
		Signature: function.Signature,

		Info: function.Info,
	}
}
