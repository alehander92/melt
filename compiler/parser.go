package compiler

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/davecgh/go-spew/spew"

	"gitlab.com/alehander42/melt/types"
)

func Parse(source string) (Module, error) {
	indented, err := Preprocess(source)
	if err != nil {
		return Module{}, err
	}
	melt := &MeltParser{Buffer: indented}
	melt.Init()
	err = melt.Parse()
	if err != nil {
		return Module{}, err
	} else {
		sexp, err := Load(melt)
		if err != nil {
			return Module{}, err
		}
		spew.Dump(err)
		return sexp, nil
	}
}

func Load(melt *MeltParser) (Module, error) {
	ast := melt.AST()
	m, err := LoadModule(ast, melt)
	return *m, err
	// ast.up.Print(melt.Buffer)
	// fmt.Println(ast.up.next.next.up.next.token32.String())
	// return Module{}, nil
}

func LoadNode(ast *node32, melt *MeltParser) (Ast, error) {
	switch Kind(ast) {
	case "Assignment":
		return LoadAssignment(ast, melt)
	case "Call":
		return LoadNode(ast.up, melt)
	case "Expression":
		return LoadNode(ast.up, melt)
	case "BuiltinCall":
		return LoadBuiltinCall(ast, melt)
	case "FunCall":
		return LoadFunCall(ast, melt)
	case "MethodCall":
		return LoadMethodCall(ast, melt)
	case "For":
		return LoadFor(ast.up, melt)
	case "On":
		return LoadOn(ast, melt)
	case "Return":
		return LoadNode(ast.up, melt)
	case "ReturnValue":
		return LoadReturnValue(ast, melt)
	case "ReturnError":
		return LoadReturnError(ast, melt)
	case "Escalator":
		return LoadEscalate(ast, melt)
	case "Simple":
		return LoadNode(ast.up, melt)
	case "Label":
		return ToLabel(melt.Buffer[ast.begin:ast.end]), nil
	case "Number":
		if rul3s[ast.up.pegRule] == "Integer" {
			a, err := strconv.ParseInt(melt.Buffer[ast.begin:ast.end], 10, 64)
			if err != nil {
				return &Module{}, err
			}
			return ToInteger(a), nil
		} else {
			a, err := strconv.ParseFloat(melt.Buffer[ast.begin:ast.end], 64)
			if err != nil {
				return &Module{}, err
			}
			return ToFloat(a), nil
		}
	case "List":
		return LoadList(ast, melt)
	case "String":
		return LoadNode(ast.up, melt)
	case "Text":
		return ToString(melt.Buffer[ast.begin:ast.end]), nil
	case "Template":
		return LoadTemplate(ast, melt)
	case "Constant":
		z := melt.Buffer[ast.begin:ast.end]
		if z == "true" || z == "false" {
			return ToBool(z == "true"), nil
		} else {
			return &Nil{}, nil
		}
	case "Error":
		a := ToLabel(melt.Buffer[ast.up.begin:ast.up.end])
		return &Error{Label: a}, nil
	case "Line":
		return LoadNode(ast.up, melt)
	case "BinaryOperation":
		node := ast.up.up
		a, err := LoadNode(node, melt)
		if err != nil {
			return &Module{}, err
		}
		op := ast.up.next.next
		operator := AddOp
		switch melt.Buffer[op.begin:op.end] {
		case "+":
			operator = AddOp
		case "-":
			operator = SubOp
		case "*":
			operator = MultOp
		case "/":
			operator = DivideOp
		}
		node = op.next.next.up
		b, err := LoadNode(node, melt)
		if err != nil {
			return &Module{}, err
		}
		return &BinaryOperation{Left: &a, Right: &b, Op: operator}, nil
	case "UnaryOperation":
		node := ast.up
		operator := PlusOp
		if melt.Buffer[node.begin:node.end] == "-" {
			operator = MinusOp
		}
		b := node.next.up
		z, err := LoadNode(b, melt)
		if err != nil {
			return &Module{}, nil
		}
		return &UnaryOperation{Op: operator, Expression: &z}, nil
	case "IndexAssignment":
		node := ast.up
		collection, err := LoadNode(node, melt)
		if err != nil {
			return &IndexAssignment{}, err
		}

		node = node.next
		index, err := LoadNode(node, melt)
		if err != nil {
			return &IndexAssignment{}, err
		}

		node = node.next.next.next
		value, err := LoadNode(node, melt)
		if err != nil {
			return &IndexAssignment{}, err
		}

		return &IndexAssignment{Collection: &collection, Index: &index, Value: &value}, nil
	case "Interface":
		label := ast.up.next
		l := ToLabel(melt.Buffer[label.begin:label.end])
		args := label.next
		t := []types.GenericVar{}
		if args != nil && Kind(args) == "GenericArgs" {
			t = LoadGenericArgs(args, melt)
			args = args.next
		}
		code := args
		methods := []InterfaceMethod{}
		vesela := []types.Method{}
		if code != nil {
			code = code.up.next.next
			for code != nil {
				if Kind(code) == "Dedent" {
					break
				} else if Kind(code) == "Newline" {
					code = code.next
					continue
				} else {
					f := code.up
					g := ToLabel(melt.Buffer[f.begin:f.end])
					node := f.next
					typez := []types.Type{}
					node.Print(melt.Buffer)
					fmt.Printf("WWWWWWWW")
					for node != nil {
						if Kind(node) == "Z" {
							typeNode := node.up.next
							h, err := LoadType(typeNode, melt)
							if err != nil {
								return &Interface{}, err
							}

							typez = append(typez, h)
						}
						node = node.next
					}
					if len(typez) == 0 {
						typez = append(typez, types.Empty{})
					}

					fmt.Printf("%s", typez)

					args := typez[:len(typez)-1]
					returnType := typez[len(typez)-1]
					er := types.Correct
					if g.Label[len(g.Label)-1] == '!' {
						er = types.Fail
					} else if g.Label[len(g.Label)-1] == '?' {
						er = types.Maybe
					}
					methods = append(methods, InterfaceMethod{Label: g, Type: types.Function{Args: args, Return: returnType, GenericVars: t, InstanceVars: make([]types.Type, len(t)), Error: er}})
					vesela = append(vesela, types.Method{Label: g.Label, Function: methods[len(methods)-1].Type})
					code = code.next
				}
			}
		}
		return &Interface{
			Label:   l,
			Info:    Info{meltType: types.NewInterface(l.Label, vesela, t)},
			Methods: methods}, nil
	case "Record":
		node := ast.up.next
		label := ToLabel(melt.Buffer[node.begin:node.end])
		node = node.next
		t := []types.GenericVar{}
		if node != nil && Kind(node) == "GenericArgs" {
			t = LoadGenericArgs(node, melt)
			node = node.next
		}
		fields := []Field{}
		typeFields := make(map[string]types.Type)
		if node != nil {
			node = node.up.next.next
			for node != nil {
				sex := node.up
				if sex == nil {
					break
				}
				fieldLabel := ToLabel(melt.Buffer[sex.begin:sex.end])
				fieldType, err := LoadType(sex.next.next, melt)

				if err != nil {
					return &Record{}, err
				}

				last := Field{Label: fieldLabel, Info: Info{meltType: fieldType}}
				fields = append(fields, last)
				typeFields[last.Label.Label] = fieldType
				node = node.next
			}
		}
		recordType := types.Record{GenericVars: t, InstanceVars: make([]types.Type, len(t)), Fields: typeFields, Label: label.Label}
		return &Record{Info: Info{meltType: recordType}, Fields: fields, Label: label}, nil
	case "Top":
		return LoadNode(ast.up, melt)
	default:
		fmt.Println("%s", Kind(ast))
	}
	return &Module{}, errors.New("wtf")
}

func LoadModule(ast *node32, melt *MeltParser) (*Module, error) {
	meltPackage := ast.up
	packageLabel := melt.Buffer[meltPackage.up.next.begin:meltPackage.up.next.end]

	next := meltPackage.next.next

	var imports MeltImport

	var err error

	if Kind(next) == "Import" {
		imports, err = LoadImport(next, melt)
		if err != nil {
			return &Module{}, err
		}
		next = next.next
		if Kind(next) == "Newline" {
			next = next.next
		}
	} else {
		imports = MeltImport{}
	}
	functions := []*Function{}
	interfaces := []*Interface{}
	records := []*Record{}
	for {
		if Kind(next) == "Newline" {
			if melt.Buffer[next.begin:next.end] == "\n\n" {
				break
			} else {
				next = next.next
				continue
			}
		}
		if next == nil {
			break
		}
		if Kind(next) != "Top" {
			next = next.next
			continue
		}
		if Kind(next.up) == "Function" {
			f, err := LoadFunction(next.up, melt)
			if err != nil {
				return &Module{}, err
			}
			functions = append(functions, &f)
		} else if Kind(next.up) == "Interface" {
			result, err := LoadNode(next.up, melt)
			if err != nil {
				return &Module{}, err
			}
			i, ok := result.(*Interface)
			if ok {
				interfaces = append(interfaces, i)
			} else {
				return &Module{}, errors.New("wat")
			}
		} else if Kind(next.up) == "Record" {
			result, err := LoadNode(next.up, melt)
			if err != nil {
				return &Module{}, err
			}
			r, ok := result.(*Record)
			if ok {
				records = append(records, r)
			} else {
				return &Module{}, errors.New("wat")
			}
		}
		next = next.next
	}
	return &Module{Package: packageLabel, Imports: &imports, Interfaces: interfaces, Records: records, Functions: functions}, nil
}

func LoadImport(ast *node32, melt *MeltParser) (MeltImport, error) {
	return MeltImport{}, nil
}

func LoadAssignment(ast *node32, melt *MeltParser) (*Set, error) {
	label := ToLabel(melt.Buffer[ast.up.begin:ast.up.end])
	value, err := LoadNode(ast.up.next.next.next, melt)
	if err != nil {
		return &Set{}, err
	}
	return &Set{Label: label, Value: &value}, nil
}

func LoadBuiltinCall(ast *node32, melt *MeltParser) (*Make, error) {
	argNode := ast.up.next
	a, err := LoadType(argNode, melt)
	if err != nil {
		return &Make{}, err
	}
	node := argNode.next
	args := []Ast{}
	for node != nil {
		if Kind(node) == "BuiltinArg" {
			arg, err := LoadNode(node.up, melt)
			if err != nil {
				return &Make{}, err
			}
			args = append(args, arg)
		}
		node = node.next
	}
	return &Make{Type: a, Args: args}, nil
}

func LoadFunCall(ast *node32, melt *MeltParser) (*Call, error) {
	label := ToLabel(melt.Buffer[ast.up.begin:ast.up.end])
	a := ast.up.next
	args := []Ast{}
	for {
		if a == nil {
			break
		} else if rul3s[a.pegRule] == "Whitespace" {
			a = a.next
			continue
		} else {
			value, err := LoadNode(a, melt)
			if err != nil {
				return &Call{}, err
			}

			args = append(args, value)
			a = a.next
		}
	}
	return &Call{Function: label, Args: args}, nil
}

func LoadMethodCall(ast *node32, melt *MeltParser) (*MethodCall, error) {
	simple := ast.up
	receiver, err := LoadNode(simple, melt)
	if err != nil {
		return &MethodCall{}, err
	}
	method := simple.next
	label := ToLabel(melt.Buffer[method.begin:method.end])
	args := []Ast{}
	a := method.next
	for {
		if a == nil {
			break
		} else if rul3s[a.pegRule] == "Whitespace" {
			a = a.next
			continue
		} else {
			arg, err := LoadNode(a, melt)
			if err != nil {
				return &MethodCall{}, err
			}
			args = append(args, arg)
			a = a.next
		}
	}
	return &MethodCall{Receiver: &receiver, Method: label, Args: args}, nil
}

func LoadFor(ast *node32, melt *MeltParser) (Ast, error) {
	a := rul3s[ast.pegRule]
	if a == "ForLoop" {
		index := ToLabel(melt.Buffer[ast.up.next.begin:ast.up.next.end])
		in := ast.up.next.next.next.next
		a := in.up
		op := a.next
		b := op.next
		begin, err := LoadNode(a, melt)
		if err != nil {
			return &ForLoop{}, err
		}
		end, err := LoadNode(b, melt)
		if err != nil {
			return &ForLoop{}, err
		}
		code := in.next.next.next
		c, err := LoadCode(code, melt)
		if err != nil {
			return &ForLoop{}, err
		}

		return &ForLoop{Index: index, Begin: &begin, End: &end, Code: &c}, nil
	} else {
		node := ast.up.next
		index := []Label{}
		for rul3s[node.pegRule] != "Expression" {
			if rul3s[node.pegRule] != "Whitespace" {
				a := ToLabel(melt.Buffer[node.begin:node.end])
				index = append(index, *a)
			}
			node = node.next
		}
		expression, err := LoadNode(node, melt)
		if err != nil {
			return &ForIn{}, err
		}
		c := node.next.next.next
		code, err := LoadCode(c, melt)
		if err != nil {
			return &ForIn{}, err
		}
		return &ForIn{Index: index, Code: &code, Sequence: &expression}, nil
	}
}

func LoadEscalate(ast *node32, melt *MeltParser) (*Escalate, error) {
	node := ast.up.next
	args := []*Label{}

	for node != nil {
		if Kind(node) != "Whitespace" {
			args = append(args, ToLabel(melt.Buffer[node.begin:node.end]))
		}
		node = node.next
	}

	return &Escalate{Args: args}, nil
}

func LoadOn(ast *node32, melt *MeltParser) (*On, error) {
	node := ast.up.next
	label := ToLabel(melt.Buffer[node.begin:node.end])
	code := node.next.next.next
	c, err := LoadCode(code, melt)
	if err != nil {
		return &On{}, err
	}

	return &On{Label: label, Handler: &c}, nil
}

func LoadReturnValue(ast *node32, melt *MeltParser) (*Return, error) {
	node := ast.up.next
	as, err := LoadNode(node, melt)
	if err != nil {
		return &Return{}, err
	}
	return &Return{Value: &as}, nil
}

func LoadReturnError(node *node32, melt *MeltParser) (*ReturnError, error) {
	if node != nil {
		node := node.up.next
		as, err := LoadNode(node, melt)
		if err != nil {
			return &ReturnError{}, err
		}
		return &ReturnError{Value: &as}, nil
	}
	return &ReturnError{}, errors.New("err")
}

func LoadCode(ast *node32, melt *MeltParser) (Code, error) {
	node := ast.up
	e := []Ast{}
	for node != nil {
		if rul3s[node.pegRule] == "Line" {
			node = node.up
			result, err := LoadNode(node, melt)
			if err != nil {
				return Code{}, err
			}
			e = append(e, result)
		}
		node = node.next
	}
	return Code{E: e}, nil
}
func LoadTemplate(ast *node32, melt *MeltParser) (*Template, error) {
	node := ast.up
	text := []string{}
	args := []Ast{}
	for node != nil && rul3s[node.pegRule] != "Q" {
		if rul3s[node.pegRule] == "Slot" {
			e := node.up
			object, err := LoadNode(e, melt)
			if err != nil {
				return &Template{}, err
			}
			args = append(args, object)
		} else {
			text = append(text, melt.Buffer[node.begin:node.end])
		}
		node = node.next
	}
	if node == nil {
		text = append(text, "")
	} else {
		text = append(text, melt.Buffer[node.begin:node.end])
	}
	return &Template{Text: text, Args: args}, nil
}

func LoadGenericArgs(ast *node32, melt *MeltParser) []types.GenericVar {
	generic := []types.GenericVar{}
	node := ast.up
	for node != nil {
		if Kind(node) != "Whitespace" {
			generic = append(generic, types.GenericVar{Label: melt.Buffer[node.begin:node.end]})
		}
		node = node.next
	}
	return generic
}

func LoadFunction(ast *node32, melt *MeltParser) (Function, error) {
	label := melt.Buffer[ast.up.next.begin:ast.up.next.end]
	args := ast.up.next.next
	functionArgs := []Arg{}
	var funArgs *Signature
	var returnType *node32
	t := []types.GenericVar{}
	if Kind(args) == "GenericArgs" {
		t = LoadGenericArgs(args, melt)
		args = args.next
	}
	if Kind(args) != "FunArgs" {
		funArgs = &Signature{Args: []types.Type{}, Return: types.Nil{}}
		returnType = args
	} else {
		arg := args.up
		funArgs = &Signature{Args: []types.Type{}}
		for {
			a := arg.up.up
			b := melt.Buffer[a.begin:a.end]
			c := a.next.next
			d := c.up
			e, err := LoadType(d, melt)
			if err != nil {
				return Function{}, err
			}
			if b[len(b)-1] == '!' || b[len(b)-1] == '?' {
				f, ok := e.(types.Function)
				if !ok {
					return Function{}, fmt.Errorf("%s is not a function", b)
				}

				if b[len(b)-1] == '!' {
					f.Error = types.Fail
				} else if b[len(b)-1] == '?' {
					f.Error = types.Maybe
				}
				e = f
				b = b[:len(b)-1]
			}

			funArgs.Args = append(funArgs.Args, e)
			functionArgs = append(functionArgs, Arg{ID: ToLabel(b), Type: e})
			if arg.next == nil {
				break
			} else {
				arg = arg.next
			}
		}
		returnType = args.next
	}
	var nodeCode *node32
	if rul3s[returnType.pegRule] == "Whitespace" {
		returnType = returnType.next
	}
	if rul3s[returnType.pegRule] != "Type" {
		funArgs.Return = types.Empty{}
		nodeCode = returnType.next.next
	} else {
		var err error
		funArgs.Return, err = LoadType(returnType, melt)
		if err != nil {
			return Function{}, err
		}
		nodeCode = returnType.next.next.next
	}
	code := &Code{E: []Ast{}}
	nodeCode = nodeCode.up
	// spew.Dump(funArgs)
	// spew.Dump(functionArgs)
	// ast.Print(melt.Buffer)

	for {
		if rul3s[nodeCode.pegRule] == "Dedent" {
			break
		} else if rul3s[nodeCode.pegRule] == "Newline" {
			nodeCode = nodeCode.next
			continue
		} else {
			z, err := LoadNode(nodeCode, melt)
			if err != nil {
				return Function{}, err
			}
			code.E = append(code.E, z)
			nodeCode = nodeCode.next
		}
	}

	er := types.Correct
	if label[len(label)-1] == '?' {
		er = types.Maybe
		label = label[:len(label)-1]
	} else if label[len(label)-1] == '!' {
		er = types.Fail
		label = label[:len(label)-1]
	}

	f := types.Function{Args: funArgs.Args, Return: funArgs.Return, Error: er, GenericVars: t, InstanceVars: make([]types.Type, len(t))}
	return Function{Label: ToLabel(label), Signature: funArgs, Info: Info{meltType: f}, Args: functionArgs, Code: code}, nil

}

func LoadList(ast *node32, melt *MeltParser) (*List, error) {
	node := ast
	next := node.up
	items := []Ast{}
	for next != nil {
		if Kind(next) == "Expression" {
			o, err := LoadNode(next.up, melt)
			if err != nil {
				return &List{}, err
			}
			items = append(items, o)
		}
		next = next.next
	}
	return &List{Elements: items}, nil
}

func LoadType(ast *node32, melt *MeltParser) (types.Type, error) {
	// node.Print(melt.Buffer)
	if Kind(ast) == "Type" || Kind(ast) == "TypeExceptFun" || Kind(ast) == "BuiltinType" || Kind(ast) == "BuiltinArg" {
		return LoadType(ast.up, melt)
	} else if Kind(ast) == "BuiltinSimple" {
		return types.Basic{Label: melt.Buffer[ast.begin:ast.end]}, nil
	} else if Kind(ast) == "BuiltinSlice" {
		t, err := LoadType(ast.up, melt)
		if err != nil {
			return types.SliceBuiltin{}, err
		}
		return types.SliceBuiltin{Element: t}, nil
	} else if Kind(ast) == "PointerType" {
		object, err := LoadType(ast.up, melt)
		if err != nil {
			return types.SliceBuiltin{}, err
		}
		p := types.Pointer{Object: object}
		return p, nil
	} else if Kind(ast) == "CapitalLabel" {
		return types.Basic{Label: melt.Buffer[ast.begin:ast.end]}, nil
	} else if Kind(ast) == "GenericType" {
		node := ast.up
		label := melt.Buffer[node.begin:node.end]
		genericVars := []types.GenericVar{}
		node = node.next
		for node != nil {
			if Kind(node) != "Whitespace" {
				a := types.GenericVar{Label: melt.Buffer[node.begin:node.end]}
				genericVars = append(genericVars, a)
			}
			node = node.next
		}
		return types.Interface{Label: label, GenericVars: genericVars}, nil
	} else if Kind(ast) == "FunType" {
		node := ast.up
		args := []types.Type{}
		for node != nil {
			if Kind(node) == "TypeExceptFun" {
				a, err := LoadType(node.up, melt)
				if err != nil {
					return types.Function{}, err
				}
				args = append(args, a)
			}
			node = node.next
		}
		returnType, args := args[len(args)-1], args[:len(args)-1]
		return types.Function{Args: args, Return: returnType, Error: types.Correct}, nil
	}
	fmt.Println("ast")
	ast.Print(melt.Buffer)
	return types.Nil{}, errors.New("No type")
}

func Kind(ast *node32) string {
	return rul3s[ast.pegRule]
}

func Preprocess(source string) (string, error) {
	lines := strings.Split(source, "\n")
	var level uint
	level = 0
	var z []string
	for a, line := range lines {
		trimmed := strings.TrimRight(strings.Trim(line, " "), "\t")
		if len(trimmed) == 0 || trimmed[0] == '#' {
			continue
		}

		new_level := IndentLevel(trimmed)
		if new_level > level+1 {
			return "", errors.New(fmt.Sprintf("line %d: indented too much\n%s", a+1, line))
		} else if new_level == level+1 {
			z = append(z, fmt.Sprintf("@@indent@@%s", line[new_level:]))
			level += 1
		} else if new_level == level {
			z = append(z, line[new_level:])
		} else {
			y := strings.Repeat("@@dedent@@\n", int(level-new_level))
			z = append(z, fmt.Sprintf("%s%s", y, line[new_level:]))
			level = new_level
		}
	}
	z = append(z, strings.Repeat("@@dedent@@\n", int(level)))
	result := strings.Join(z, "\n") + "\n"
	fmt.Println(result)
	return result, nil
}

func IndentLevel(line string) uint {
	for a, c := range line {
		if c != '\t' {
			return uint(a)
		}
	}
	return 0
}

//Parse melt to ast
// func Parse2(source string) (Ast, error) {
// 	value := &Return{Value: &BinaryOperation{
// 		Op:   MultOp,
// 		Left: ToLabel("n"),
// 		Right: &Call{
// 			Function: ToLabel("fac"),
// 			Args: []Ast{BinaryOperation{
// 				Op:    SubOp,
// 				Left:  ToLabel("n"),
// 				Right: ToInteger(1)}}}}}

// 	code := &Code{E: []Ast{&If{
// 		Test: &Cmp{
// 			Op:    EqualOp,
// 			Left:  ToLabel("n"),
// 			Right: ToInteger(1)},
// 		Code: &Code{
// 			E: []Ast{
// 				&Return{Value: ToInteger(1)}}},
// 		Otherwise: &Code{E: []Ast{value}}}}}

// 	return Module{
// 		Imports: MeltImport{
// 			Go:   []Import{},
// 			Melt: []Import{}},
// 		Functions: []Function{{
// 			Label: ToLabel("fac"),
// 			Args:  []Arg{{ID: ToLabel("n"), Type: ToType("int")}},
// 			Signature: &Signature{
// 				Args:   []Type{ToType("int")},
// 				Return: ToType("int")},
// 			Code: code}}}, nil
// }
