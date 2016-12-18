package compiler

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

const endSymbol rune = 1114112

/* The rule types inferred from the grammar are below. */
type pegRule uint8

const (
	ruleUnknown pegRule = iota
	ruleModule
	ruleTop
	rulePackage
	ruleImport
	ruleGoImport
	ruleMeltImport
	ruleFunction
	ruleInterface
	ruleArray
	ruleDeclaration
	ruleZ
	ruleRecord
	ruleRecordContents
	ruleSex
	ruleFunArgs
	ruleGenericArgs
	ruleFunArg
	rulePreArg
	ruleLastArg
	ruleFunLabel
	ruleType
	rulePointerType
	ruleFunType
	ruleGenericType
	ruleBuiltinType
	ruleBuiltinSimple
	ruleBuiltinSlice
	ruleBuiltinArray
	ruleBuiltinMap
	ruleTypeExceptFun
	ruleCode
	ruleIndent
	ruleDedent
	ruleLine
	ruleIndexAssignment
	ruleAssignment
	ruleExpression
	ruleSimple
	ruleList
	ruleBinaryOperation
	ruleExpressionExceptBinaryOperation
	ruleBinaryOperator
	ruleUnaryOperation
	ruleUnaryOperator
	ruleExpressionExceptOperation
	ruleMethodCall
	ruleCall
	ruleBuiltinCall
	ruleBuiltinFun
	ruleBuiltinArg
	ruleFunCall
	ruleFor
	ruleForIn
	ruleForLoop
	ruleRange
	ruleRangeOperator
	ruleOn
	ruleReturn
	ruleReturnValue
	ruleReturnError
	ruleEscalator
	ruleLowerLabel
	ruleCapitalLabel
	ruleFunLowerLabel
	ruleLabel
	ruleFloat
	ruleInteger
	ruleNumber
	ruleConstant
	ruleString
	ruleTemplate
	ruleSegment
	ruleQ
	ruleText
	ruleError
	ruleSlot
	ruleWhitespace
	ruleNewline
	ruleEOT
)

var rul3s = [...]string{
	"Unknown",
	"Module",
	"Top",
	"Package",
	"Import",
	"GoImport",
	"MeltImport",
	"Function",
	"Interface",
	"Array",
	"Declaration",
	"Z",
	"Record",
	"RecordContents",
	"Sex",
	"FunArgs",
	"GenericArgs",
	"FunArg",
	"PreArg",
	"LastArg",
	"FunLabel",
	"Type",
	"PointerType",
	"FunType",
	"GenericType",
	"BuiltinType",
	"BuiltinSimple",
	"BuiltinSlice",
	"BuiltinArray",
	"BuiltinMap",
	"TypeExceptFun",
	"Code",
	"Indent",
	"Dedent",
	"Line",
	"IndexAssignment",
	"Assignment",
	"Expression",
	"Simple",
	"List",
	"BinaryOperation",
	"ExpressionExceptBinaryOperation",
	"BinaryOperator",
	"UnaryOperation",
	"UnaryOperator",
	"ExpressionExceptOperation",
	"MethodCall",
	"Call",
	"BuiltinCall",
	"BuiltinFun",
	"BuiltinArg",
	"FunCall",
	"For",
	"ForIn",
	"ForLoop",
	"Range",
	"RangeOperator",
	"On",
	"Return",
	"ReturnValue",
	"ReturnError",
	"Escalator",
	"LowerLabel",
	"CapitalLabel",
	"FunLowerLabel",
	"Label",
	"Float",
	"Integer",
	"Number",
	"Constant",
	"String",
	"Template",
	"Segment",
	"Q",
	"Text",
	"Error",
	"Slot",
	"Whitespace",
	"Newline",
	"EOT",
}

type token32 struct {
	pegRule
	begin, end uint32
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v", rul3s[t.pegRule], t.begin, t.end)
}

type node32 struct {
	token32
	up, next *node32
}

func (node *node32) Print(buffer string) {
	var print func(node *node32, depth int)
	print = func(node *node32, depth int) {
		for node != nil {
			for c := 0; c < depth; c++ {
				fmt.Printf(" ")
			}
			fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[node.pegRule], strconv.Quote(string(([]rune(buffer)[node.begin:node.end]))))
			if node.up != nil {
				print(node.up, depth+1)
			}
			node = node.next
		}
	}
	print(node, 0)
}

type tokens32 struct {
	tree []token32
}

func (t *tokens32) Trim(length uint32) {
	t.tree = t.tree[:length]
}

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens32) AST() *node32 {
	type element struct {
		node *node32
		down *element
	}
	tokens := t.Tokens()
	var stack *element
	for _, token := range tokens {
		if token.begin == token.end {
			continue
		}
		node := &node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	if stack != nil {
		return stack.node
	}
	return nil
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	t.AST().Print(buffer)
}

func (t *tokens32) Add(rule pegRule, begin, end, index uint32) {
	if tree := t.tree; int(index) >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		copy(expanded, tree)
		t.tree = expanded
	}
	t.tree[index] = token32{
		pegRule: rule,
		begin:   begin,
		end:     end,
	}
}

func (t *tokens32) Tokens() []token32 {
	return t.tree
}

type MeltParser struct {
	Buffer string
	buffer []rune
	rules  [80]func() bool
	parse  func(rule ...int) error
	reset  func()
	Pretty bool
	tokens32
}

func (p *MeltParser) Parse(rule ...int) error {
	return p.parse(rule...)
}

func (p *MeltParser) Reset() {
	p.reset()
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer []rune, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range buffer {
		if c == '\n' {
			line, symbol = line+1, 0
		} else {
			symbol++
		}
		if i == positions[j] {
			translations[positions[j]] = textPosition{line, symbol}
			for j++; j < length; j++ {
				if i != positions[j] {
					continue search
				}
			}
			break search
		}
	}

	return translations
}

type parseError struct {
	p   *MeltParser
	max token32
}

func (e *parseError) Error() string {
	tokens, error := []token32{e.max}, "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.buffer, positions)
	format := "parse error near %v (line %v symbol %v - line %v symbol %v):\n%v\n"
	if e.p.Pretty {
		format = "parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n"
	}
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		error += fmt.Sprintf(format,
			rul3s[token.pegRule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			strconv.Quote(string(e.p.buffer[begin:end])))
	}

	return error
}

func (p *MeltParser) PrintSyntaxTree() {
	p.tokens32.PrintSyntaxTree(p.Buffer)
}

func (p *MeltParser) Init() {
	var (
		max                  token32
		position, tokenIndex uint32
		buffer               []rune
	)
	p.reset = func() {
		max = token32{}
		position, tokenIndex = 0, 0

		p.buffer = []rune(p.Buffer)
		if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != endSymbol {
			p.buffer = append(p.buffer, endSymbol)
		}
		buffer = p.buffer
	}
	p.reset()

	_rules, tree := p.rules, tokens32{tree: make([]token32, math.MaxInt16)}
	p.parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.tokens32 = tree
		if matches {
			p.Trim(tokenIndex)
			return nil
		}
		return &parseError{p, max}
	}

	add := func(rule pegRule, begin uint32) {
		tree.Add(rule, begin, position, tokenIndex)
		tokenIndex++
		if begin != position && position > max.end {
			max = token32{rule, begin, position}
		}
	}

	matchDot := func() bool {
		if buffer[position] != endSymbol {
			position++
			return true
		}
		return false
	}

	/*matchChar := func(c byte) bool {
		if buffer[position] == c {
			position++
			return true
		}
		return false
	}*/

	/*matchRange := func(lower byte, upper byte) bool {
		if c := buffer[position]; c >= lower && c <= upper {
			position++
			return true
		}
		return false
	}*/

	_rules = [...]func() bool{
		nil,
		/* 0 Module <- <(Package Newline Import? Newline? (Top Newline)* EOT)> */
		func() bool {
			position0, tokenIndex0 := position, tokenIndex
			{
				position1 := position
				{
					position2 := position
					{
						position3, tokenIndex3 := position, tokenIndex
						if buffer[position] != rune('p') {
							goto l4
						}
						position++
						goto l3
					l4:
						position, tokenIndex = position3, tokenIndex3
						if buffer[position] != rune('P') {
							goto l0
						}
						position++
					}
				l3:
					{
						position5, tokenIndex5 := position, tokenIndex
						if buffer[position] != rune('a') {
							goto l6
						}
						position++
						goto l5
					l6:
						position, tokenIndex = position5, tokenIndex5
						if buffer[position] != rune('A') {
							goto l0
						}
						position++
					}
				l5:
					{
						position7, tokenIndex7 := position, tokenIndex
						if buffer[position] != rune('c') {
							goto l8
						}
						position++
						goto l7
					l8:
						position, tokenIndex = position7, tokenIndex7
						if buffer[position] != rune('C') {
							goto l0
						}
						position++
					}
				l7:
					{
						position9, tokenIndex9 := position, tokenIndex
						if buffer[position] != rune('k') {
							goto l10
						}
						position++
						goto l9
					l10:
						position, tokenIndex = position9, tokenIndex9
						if buffer[position] != rune('K') {
							goto l0
						}
						position++
					}
				l9:
					{
						position11, tokenIndex11 := position, tokenIndex
						if buffer[position] != rune('a') {
							goto l12
						}
						position++
						goto l11
					l12:
						position, tokenIndex = position11, tokenIndex11
						if buffer[position] != rune('A') {
							goto l0
						}
						position++
					}
				l11:
					{
						position13, tokenIndex13 := position, tokenIndex
						if buffer[position] != rune('g') {
							goto l14
						}
						position++
						goto l13
					l14:
						position, tokenIndex = position13, tokenIndex13
						if buffer[position] != rune('G') {
							goto l0
						}
						position++
					}
				l13:
					{
						position15, tokenIndex15 := position, tokenIndex
						if buffer[position] != rune('e') {
							goto l16
						}
						position++
						goto l15
					l16:
						position, tokenIndex = position15, tokenIndex15
						if buffer[position] != rune('E') {
							goto l0
						}
						position++
					}
				l15:
					if !_rules[ruleWhitespace]() {
						goto l0
					}
					if !_rules[ruleLowerLabel]() {
						goto l0
					}
					add(rulePackage, position2)
				}
				if !_rules[ruleNewline]() {
					goto l0
				}
				{
					position17, tokenIndex17 := position, tokenIndex
					{
						position19 := position
						{
							position20, tokenIndex20 := position, tokenIndex
							if buffer[position] != rune('i') {
								goto l21
							}
							position++
							goto l20
						l21:
							position, tokenIndex = position20, tokenIndex20
							if buffer[position] != rune('I') {
								goto l17
							}
							position++
						}
					l20:
						{
							position22, tokenIndex22 := position, tokenIndex
							if buffer[position] != rune('m') {
								goto l23
							}
							position++
							goto l22
						l23:
							position, tokenIndex = position22, tokenIndex22
							if buffer[position] != rune('M') {
								goto l17
							}
							position++
						}
					l22:
						{
							position24, tokenIndex24 := position, tokenIndex
							if buffer[position] != rune('p') {
								goto l25
							}
							position++
							goto l24
						l25:
							position, tokenIndex = position24, tokenIndex24
							if buffer[position] != rune('P') {
								goto l17
							}
							position++
						}
					l24:
						{
							position26, tokenIndex26 := position, tokenIndex
							if buffer[position] != rune('o') {
								goto l27
							}
							position++
							goto l26
						l27:
							position, tokenIndex = position26, tokenIndex26
							if buffer[position] != rune('O') {
								goto l17
							}
							position++
						}
					l26:
						{
							position28, tokenIndex28 := position, tokenIndex
							if buffer[position] != rune('r') {
								goto l29
							}
							position++
							goto l28
						l29:
							position, tokenIndex = position28, tokenIndex28
							if buffer[position] != rune('R') {
								goto l17
							}
							position++
						}
					l28:
						{
							position30, tokenIndex30 := position, tokenIndex
							if buffer[position] != rune('t') {
								goto l31
							}
							position++
							goto l30
						l31:
							position, tokenIndex = position30, tokenIndex30
							if buffer[position] != rune('T') {
								goto l17
							}
							position++
						}
					l30:
						if buffer[position] != rune(':') {
							goto l17
						}
						position++
						if !_rules[ruleNewline]() {
							goto l17
						}
						if !_rules[ruleIndent]() {
							goto l17
						}
						{
							position32, tokenIndex32 := position, tokenIndex
							{
								position34 := position
								{
									position35, tokenIndex35 := position, tokenIndex
									if buffer[position] != rune('g') {
										goto l36
									}
									position++
									goto l35
								l36:
									position, tokenIndex = position35, tokenIndex35
									if buffer[position] != rune('G') {
										goto l32
									}
									position++
								}
							l35:
								{
									position37, tokenIndex37 := position, tokenIndex
									if buffer[position] != rune('o') {
										goto l38
									}
									position++
									goto l37
								l38:
									position, tokenIndex = position37, tokenIndex37
									if buffer[position] != rune('O') {
										goto l32
									}
									position++
								}
							l37:
								if buffer[position] != rune(':') {
									goto l32
								}
								position++
								if !_rules[ruleNewline]() {
									goto l32
								}
								if !_rules[ruleIndent]() {
									goto l32
								}
								if !_rules[ruleString]() {
									goto l32
								}
								if !_rules[ruleNewline]() {
									goto l32
								}
							l39:
								{
									position40, tokenIndex40 := position, tokenIndex
									if !_rules[ruleString]() {
										goto l40
									}
									if !_rules[ruleNewline]() {
										goto l40
									}
									goto l39
								l40:
									position, tokenIndex = position40, tokenIndex40
								}
								if !_rules[ruleDedent]() {
									goto l32
								}
								add(ruleGoImport, position34)
							}
							goto l33
						l32:
							position, tokenIndex = position32, tokenIndex32
						}
					l33:
						{
							position41, tokenIndex41 := position, tokenIndex
							{
								position43 := position
								{
									position44, tokenIndex44 := position, tokenIndex
									if buffer[position] != rune('m') {
										goto l45
									}
									position++
									goto l44
								l45:
									position, tokenIndex = position44, tokenIndex44
									if buffer[position] != rune('M') {
										goto l41
									}
									position++
								}
							l44:
								{
									position46, tokenIndex46 := position, tokenIndex
									if buffer[position] != rune('e') {
										goto l47
									}
									position++
									goto l46
								l47:
									position, tokenIndex = position46, tokenIndex46
									if buffer[position] != rune('E') {
										goto l41
									}
									position++
								}
							l46:
								{
									position48, tokenIndex48 := position, tokenIndex
									if buffer[position] != rune('l') {
										goto l49
									}
									position++
									goto l48
								l49:
									position, tokenIndex = position48, tokenIndex48
									if buffer[position] != rune('L') {
										goto l41
									}
									position++
								}
							l48:
								{
									position50, tokenIndex50 := position, tokenIndex
									if buffer[position] != rune('t') {
										goto l51
									}
									position++
									goto l50
								l51:
									position, tokenIndex = position50, tokenIndex50
									if buffer[position] != rune('T') {
										goto l41
									}
									position++
								}
							l50:
								if buffer[position] != rune(':') {
									goto l41
								}
								position++
								if !_rules[ruleNewline]() {
									goto l41
								}
								if !_rules[ruleIndent]() {
									goto l41
								}
								if !_rules[ruleString]() {
									goto l41
								}
								if !_rules[ruleNewline]() {
									goto l41
								}
							l52:
								{
									position53, tokenIndex53 := position, tokenIndex
									if !_rules[ruleString]() {
										goto l53
									}
									if !_rules[ruleNewline]() {
										goto l53
									}
									goto l52
								l53:
									position, tokenIndex = position53, tokenIndex53
								}
								if !_rules[ruleDedent]() {
									goto l41
								}
								add(ruleMeltImport, position43)
							}
							goto l42
						l41:
							position, tokenIndex = position41, tokenIndex41
						}
					l42:
						add(ruleImport, position19)
					}
					goto l18
				l17:
					position, tokenIndex = position17, tokenIndex17
				}
			l18:
				{
					position54, tokenIndex54 := position, tokenIndex
					if !_rules[ruleNewline]() {
						goto l54
					}
					goto l55
				l54:
					position, tokenIndex = position54, tokenIndex54
				}
			l55:
			l56:
				{
					position57, tokenIndex57 := position, tokenIndex
					{
						position58 := position
						{
							switch buffer[position] {
							case 'R', 'r':
								{
									position60 := position
									{
										position61, tokenIndex61 := position, tokenIndex
										if buffer[position] != rune('r') {
											goto l62
										}
										position++
										goto l61
									l62:
										position, tokenIndex = position61, tokenIndex61
										if buffer[position] != rune('R') {
											goto l57
										}
										position++
									}
								l61:
									{
										position63, tokenIndex63 := position, tokenIndex
										if buffer[position] != rune('e') {
											goto l64
										}
										position++
										goto l63
									l64:
										position, tokenIndex = position63, tokenIndex63
										if buffer[position] != rune('E') {
											goto l57
										}
										position++
									}
								l63:
									{
										position65, tokenIndex65 := position, tokenIndex
										if buffer[position] != rune('c') {
											goto l66
										}
										position++
										goto l65
									l66:
										position, tokenIndex = position65, tokenIndex65
										if buffer[position] != rune('C') {
											goto l57
										}
										position++
									}
								l65:
									{
										position67, tokenIndex67 := position, tokenIndex
										if buffer[position] != rune('o') {
											goto l68
										}
										position++
										goto l67
									l68:
										position, tokenIndex = position67, tokenIndex67
										if buffer[position] != rune('O') {
											goto l57
										}
										position++
									}
								l67:
									{
										position69, tokenIndex69 := position, tokenIndex
										if buffer[position] != rune('r') {
											goto l70
										}
										position++
										goto l69
									l70:
										position, tokenIndex = position69, tokenIndex69
										if buffer[position] != rune('R') {
											goto l57
										}
										position++
									}
								l69:
									{
										position71, tokenIndex71 := position, tokenIndex
										if buffer[position] != rune('d') {
											goto l72
										}
										position++
										goto l71
									l72:
										position, tokenIndex = position71, tokenIndex71
										if buffer[position] != rune('D') {
											goto l57
										}
										position++
									}
								l71:
									if !_rules[ruleWhitespace]() {
										goto l57
									}
									if !_rules[ruleCapitalLabel]() {
										goto l57
									}
									{
										position73, tokenIndex73 := position, tokenIndex
										if !_rules[ruleGenericArgs]() {
											goto l73
										}
										goto l74
									l73:
										position, tokenIndex = position73, tokenIndex73
									}
								l74:
									{
										position75, tokenIndex75 := position, tokenIndex
										{
											position77 := position
											if buffer[position] != rune(':') {
												goto l75
											}
											position++
											if !_rules[ruleNewline]() {
												goto l75
											}
											if !_rules[ruleIndent]() {
												goto l75
											}
											{
												position80 := position
												if !_rules[ruleLabel]() {
													goto l75
												}
												if !_rules[ruleWhitespace]() {
													goto l75
												}
												if !_rules[ruleType]() {
													goto l75
												}
												add(ruleSex, position80)
											}
											if !_rules[ruleNewline]() {
												goto l75
											}
										l78:
											{
												position79, tokenIndex79 := position, tokenIndex
												{
													position81 := position
													if !_rules[ruleLabel]() {
														goto l79
													}
													if !_rules[ruleWhitespace]() {
														goto l79
													}
													if !_rules[ruleType]() {
														goto l79
													}
													add(ruleSex, position81)
												}
												if !_rules[ruleNewline]() {
													goto l79
												}
												goto l78
											l79:
												position, tokenIndex = position79, tokenIndex79
											}
											if !_rules[ruleDedent]() {
												goto l75
											}
											add(ruleRecordContents, position77)
										}
										goto l76
									l75:
										position, tokenIndex = position75, tokenIndex75
									}
								l76:
									add(ruleRecord, position60)
								}
								break
							case 'I', 'i':
								{
									position82 := position
									{
										position83, tokenIndex83 := position, tokenIndex
										if buffer[position] != rune('i') {
											goto l84
										}
										position++
										goto l83
									l84:
										position, tokenIndex = position83, tokenIndex83
										if buffer[position] != rune('I') {
											goto l57
										}
										position++
									}
								l83:
									{
										position85, tokenIndex85 := position, tokenIndex
										if buffer[position] != rune('n') {
											goto l86
										}
										position++
										goto l85
									l86:
										position, tokenIndex = position85, tokenIndex85
										if buffer[position] != rune('N') {
											goto l57
										}
										position++
									}
								l85:
									{
										position87, tokenIndex87 := position, tokenIndex
										if buffer[position] != rune('t') {
											goto l88
										}
										position++
										goto l87
									l88:
										position, tokenIndex = position87, tokenIndex87
										if buffer[position] != rune('T') {
											goto l57
										}
										position++
									}
								l87:
									{
										position89, tokenIndex89 := position, tokenIndex
										if buffer[position] != rune('e') {
											goto l90
										}
										position++
										goto l89
									l90:
										position, tokenIndex = position89, tokenIndex89
										if buffer[position] != rune('E') {
											goto l57
										}
										position++
									}
								l89:
									{
										position91, tokenIndex91 := position, tokenIndex
										if buffer[position] != rune('r') {
											goto l92
										}
										position++
										goto l91
									l92:
										position, tokenIndex = position91, tokenIndex91
										if buffer[position] != rune('R') {
											goto l57
										}
										position++
									}
								l91:
									{
										position93, tokenIndex93 := position, tokenIndex
										if buffer[position] != rune('f') {
											goto l94
										}
										position++
										goto l93
									l94:
										position, tokenIndex = position93, tokenIndex93
										if buffer[position] != rune('F') {
											goto l57
										}
										position++
									}
								l93:
									{
										position95, tokenIndex95 := position, tokenIndex
										if buffer[position] != rune('a') {
											goto l96
										}
										position++
										goto l95
									l96:
										position, tokenIndex = position95, tokenIndex95
										if buffer[position] != rune('A') {
											goto l57
										}
										position++
									}
								l95:
									{
										position97, tokenIndex97 := position, tokenIndex
										if buffer[position] != rune('c') {
											goto l98
										}
										position++
										goto l97
									l98:
										position, tokenIndex = position97, tokenIndex97
										if buffer[position] != rune('C') {
											goto l57
										}
										position++
									}
								l97:
									{
										position99, tokenIndex99 := position, tokenIndex
										if buffer[position] != rune('e') {
											goto l100
										}
										position++
										goto l99
									l100:
										position, tokenIndex = position99, tokenIndex99
										if buffer[position] != rune('E') {
											goto l57
										}
										position++
									}
								l99:
									if !_rules[ruleWhitespace]() {
										goto l57
									}
									if !_rules[ruleCapitalLabel]() {
										goto l57
									}
									{
										position101, tokenIndex101 := position, tokenIndex
										if !_rules[ruleGenericArgs]() {
											goto l101
										}
										goto l102
									l101:
										position, tokenIndex = position101, tokenIndex101
									}
								l102:
									{
										position103, tokenIndex103 := position, tokenIndex
										{
											position105 := position
											if buffer[position] != rune(':') {
												goto l103
											}
											position++
											if !_rules[ruleNewline]() {
												goto l103
											}
											if !_rules[ruleIndent]() {
												goto l103
											}
											{
												position108 := position
												if !_rules[ruleFunLabel]() {
													goto l103
												}
												if buffer[position] != rune('(') {
													goto l103
												}
												position++
											l109:
												{
													position110, tokenIndex110 := position, tokenIndex
													if !_rules[ruleType]() {
														goto l110
													}
													if buffer[position] != rune(',') {
														goto l110
													}
													position++
													{
														position111, tokenIndex111 := position, tokenIndex
														if !_rules[ruleWhitespace]() {
															goto l111
														}
														goto l112
													l111:
														position, tokenIndex = position111, tokenIndex111
													}
												l112:
													goto l109
												l110:
													position, tokenIndex = position110, tokenIndex110
												}
												{
													position113, tokenIndex113 := position, tokenIndex
													if !_rules[ruleType]() {
														goto l113
													}
													goto l114
												l113:
													position, tokenIndex = position113, tokenIndex113
												}
											l114:
												if buffer[position] != rune(')') {
													goto l103
												}
												position++
												{
													position115, tokenIndex115 := position, tokenIndex
													{
														position117 := position
														if !_rules[ruleWhitespace]() {
															goto l115
														}
														if !_rules[ruleType]() {
															goto l115
														}
														add(ruleZ, position117)
													}
													goto l116
												l115:
													position, tokenIndex = position115, tokenIndex115
												}
											l116:
												add(ruleDeclaration, position108)
											}
											if !_rules[ruleNewline]() {
												goto l103
											}
										l106:
											{
												position107, tokenIndex107 := position, tokenIndex
												{
													position118 := position
													if !_rules[ruleFunLabel]() {
														goto l107
													}
													if buffer[position] != rune('(') {
														goto l107
													}
													position++
												l119:
													{
														position120, tokenIndex120 := position, tokenIndex
														if !_rules[ruleType]() {
															goto l120
														}
														if buffer[position] != rune(',') {
															goto l120
														}
														position++
														{
															position121, tokenIndex121 := position, tokenIndex
															if !_rules[ruleWhitespace]() {
																goto l121
															}
															goto l122
														l121:
															position, tokenIndex = position121, tokenIndex121
														}
													l122:
														goto l119
													l120:
														position, tokenIndex = position120, tokenIndex120
													}
													{
														position123, tokenIndex123 := position, tokenIndex
														if !_rules[ruleType]() {
															goto l123
														}
														goto l124
													l123:
														position, tokenIndex = position123, tokenIndex123
													}
												l124:
													if buffer[position] != rune(')') {
														goto l107
													}
													position++
													{
														position125, tokenIndex125 := position, tokenIndex
														{
															position127 := position
															if !_rules[ruleWhitespace]() {
																goto l125
															}
															if !_rules[ruleType]() {
																goto l125
															}
															add(ruleZ, position127)
														}
														goto l126
													l125:
														position, tokenIndex = position125, tokenIndex125
													}
												l126:
													add(ruleDeclaration, position118)
												}
												if !_rules[ruleNewline]() {
													goto l107
												}
												goto l106
											l107:
												position, tokenIndex = position107, tokenIndex107
											}
											if !_rules[ruleDedent]() {
												goto l103
											}
											add(ruleArray, position105)
										}
										goto l104
									l103:
										position, tokenIndex = position103, tokenIndex103
									}
								l104:
									add(ruleInterface, position82)
								}
								break
							default:
								{
									position128 := position
									{
										position129, tokenIndex129 := position, tokenIndex
										if buffer[position] != rune('f') {
											goto l130
										}
										position++
										goto l129
									l130:
										position, tokenIndex = position129, tokenIndex129
										if buffer[position] != rune('F') {
											goto l57
										}
										position++
									}
								l129:
									{
										position131, tokenIndex131 := position, tokenIndex
										if buffer[position] != rune('u') {
											goto l132
										}
										position++
										goto l131
									l132:
										position, tokenIndex = position131, tokenIndex131
										if buffer[position] != rune('U') {
											goto l57
										}
										position++
									}
								l131:
									{
										position133, tokenIndex133 := position, tokenIndex
										if buffer[position] != rune('n') {
											goto l134
										}
										position++
										goto l133
									l134:
										position, tokenIndex = position133, tokenIndex133
										if buffer[position] != rune('N') {
											goto l57
										}
										position++
									}
								l133:
									{
										position135, tokenIndex135 := position, tokenIndex
										if buffer[position] != rune('c') {
											goto l136
										}
										position++
										goto l135
									l136:
										position, tokenIndex = position135, tokenIndex135
										if buffer[position] != rune('C') {
											goto l57
										}
										position++
									}
								l135:
									if !_rules[ruleWhitespace]() {
										goto l57
									}
									if !_rules[ruleFunLabel]() {
										goto l57
									}
									{
										position137, tokenIndex137 := position, tokenIndex
										if !_rules[ruleGenericArgs]() {
											goto l137
										}
										goto l138
									l137:
										position, tokenIndex = position137, tokenIndex137
									}
								l138:
									{
										position139, tokenIndex139 := position, tokenIndex
										{
											position141 := position
											if buffer[position] != rune('(') {
												goto l139
											}
											position++
										l142:
											{
												position143, tokenIndex143 := position, tokenIndex
												{
													position144 := position
													{
														position145, tokenIndex145 := position, tokenIndex
														{
															position147 := position
															if !_rules[ruleFunLowerLabel]() {
																goto l146
															}
															if !_rules[ruleWhitespace]() {
																goto l146
															}
															if !_rules[ruleType]() {
																goto l146
															}
															if buffer[position] != rune(',') {
																goto l146
															}
															position++
															{
																position148, tokenIndex148 := position, tokenIndex
																if !_rules[ruleWhitespace]() {
																	goto l148
																}
																goto l149
															l148:
																position, tokenIndex = position148, tokenIndex148
															}
														l149:
															add(rulePreArg, position147)
														}
														goto l145
													l146:
														position, tokenIndex = position145, tokenIndex145
														{
															position150 := position
															if !_rules[ruleFunLowerLabel]() {
																goto l143
															}
															if !_rules[ruleWhitespace]() {
																goto l143
															}
															if !_rules[ruleType]() {
																goto l143
															}
															add(ruleLastArg, position150)
														}
													}
												l145:
													add(ruleFunArg, position144)
												}
												goto l142
											l143:
												position, tokenIndex = position143, tokenIndex143
											}
											if buffer[position] != rune(')') {
												goto l139
											}
											position++
											add(ruleFunArgs, position141)
										}
										goto l140
									l139:
										position, tokenIndex = position139, tokenIndex139
									}
								l140:
									{
										position151, tokenIndex151 := position, tokenIndex
										if !_rules[ruleWhitespace]() {
											goto l151
										}
										goto l152
									l151:
										position, tokenIndex = position151, tokenIndex151
									}
								l152:
									{
										position153, tokenIndex153 := position, tokenIndex
										if !_rules[ruleType]() {
											goto l153
										}
										goto l154
									l153:
										position, tokenIndex = position153, tokenIndex153
									}
								l154:
									if buffer[position] != rune(':') {
										goto l57
									}
									position++
									if !_rules[ruleNewline]() {
										goto l57
									}
									if !_rules[ruleIndent]() {
										goto l57
									}
									if !_rules[ruleCode]() {
										goto l57
									}
									add(ruleFunction, position128)
								}
								break
							}
						}

						add(ruleTop, position58)
					}
					if !_rules[ruleNewline]() {
						goto l57
					}
					goto l56
				l57:
					position, tokenIndex = position57, tokenIndex57
				}
				{
					position155 := position
					{
						position156, tokenIndex156 := position, tokenIndex
						if !matchDot() {
							goto l156
						}
						goto l0
					l156:
						position, tokenIndex = position156, tokenIndex156
					}
					add(ruleEOT, position155)
				}
				add(ruleModule, position1)
			}
			return true
		l0:
			position, tokenIndex = position0, tokenIndex0
			return false
		},
		/* 1 Top <- <((&('R' | 'r') Record) | (&('I' | 'i') Interface) | (&('F' | 'f') Function))> */
		nil,
		/* 2 Package <- <(('p' / 'P') ('a' / 'A') ('c' / 'C') ('k' / 'K') ('a' / 'A') ('g' / 'G') ('e' / 'E') Whitespace LowerLabel)> */
		nil,
		/* 3 Import <- <(('i' / 'I') ('m' / 'M') ('p' / 'P') ('o' / 'O') ('r' / 'R') ('t' / 'T') ':' Newline Indent GoImport? MeltImport?)> */
		nil,
		/* 4 GoImport <- <(('g' / 'G') ('o' / 'O') ':' Newline Indent (String Newline)+ Dedent)> */
		nil,
		/* 5 MeltImport <- <(('m' / 'M') ('e' / 'E') ('l' / 'L') ('t' / 'T') ':' Newline Indent (String Newline)+ Dedent)> */
		nil,
		/* 6 Function <- <(('f' / 'F') ('u' / 'U') ('n' / 'N') ('c' / 'C') Whitespace FunLabel GenericArgs? FunArgs? Whitespace? Type? ':' Newline Indent Code)> */
		nil,
		/* 7 Interface <- <(('i' / 'I') ('n' / 'N') ('t' / 'T') ('e' / 'E') ('r' / 'R') ('f' / 'F') ('a' / 'A') ('c' / 'C') ('e' / 'E') Whitespace CapitalLabel GenericArgs? Array?)> */
		nil,
		/* 8 Array <- <(':' Newline Indent (Declaration Newline)+ Dedent)> */
		nil,
		/* 9 Declaration <- <(FunLabel '(' (Type ',' Whitespace?)* Type? ')' Z?)> */
		nil,
		/* 10 Z <- <(Whitespace Type)> */
		nil,
		/* 11 Record <- <(('r' / 'R') ('e' / 'E') ('c' / 'C') ('o' / 'O') ('r' / 'R') ('d' / 'D') Whitespace CapitalLabel GenericArgs? RecordContents?)> */
		nil,
		/* 12 RecordContents <- <(':' Newline Indent (Sex Newline)+ Dedent)> */
		nil,
		/* 13 Sex <- <(Label Whitespace Type)> */
		nil,
		/* 14 FunArgs <- <('(' FunArg* ')')> */
		nil,
		/* 15 GenericArgs <- <('<' (CapitalLabel ',' Whitespace?)* CapitalLabel '>')> */
		func() bool {
			position171, tokenIndex171 := position, tokenIndex
			{
				position172 := position
				if buffer[position] != rune('<') {
					goto l171
				}
				position++
			l173:
				{
					position174, tokenIndex174 := position, tokenIndex
					if !_rules[ruleCapitalLabel]() {
						goto l174
					}
					if buffer[position] != rune(',') {
						goto l174
					}
					position++
					{
						position175, tokenIndex175 := position, tokenIndex
						if !_rules[ruleWhitespace]() {
							goto l175
						}
						goto l176
					l175:
						position, tokenIndex = position175, tokenIndex175
					}
				l176:
					goto l173
				l174:
					position, tokenIndex = position174, tokenIndex174
				}
				if !_rules[ruleCapitalLabel]() {
					goto l171
				}
				if buffer[position] != rune('>') {
					goto l171
				}
				position++
				add(ruleGenericArgs, position172)
			}
			return true
		l171:
			position, tokenIndex = position171, tokenIndex171
			return false
		},
		/* 16 FunArg <- <(PreArg / LastArg)> */
		nil,
		/* 17 PreArg <- <(FunLowerLabel Whitespace Type ',' Whitespace?)> */
		nil,
		/* 18 LastArg <- <(FunLowerLabel Whitespace Type)> */
		nil,
		/* 19 FunLabel <- <(([A-Z] / [a-z]) ((&('_') '_') | (&('`') '`') | (&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') [0-9]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]))* ('?' / '!')?)> */
		func() bool {
			position180, tokenIndex180 := position, tokenIndex
			{
				position181 := position
				{
					position182, tokenIndex182 := position, tokenIndex
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l183
					}
					position++
					goto l182
				l183:
					position, tokenIndex = position182, tokenIndex182
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l180
					}
					position++
				}
			l182:
			l184:
				{
					position185, tokenIndex185 := position, tokenIndex
					{
						switch buffer[position] {
						case '_':
							if buffer[position] != rune('_') {
								goto l185
							}
							position++
							break
						case '`':
							if buffer[position] != rune('`') {
								goto l185
							}
							position++
							break
						case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l185
							}
							position++
							break
						case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z':
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l185
							}
							position++
							break
						default:
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l185
							}
							position++
							break
						}
					}

					goto l184
				l185:
					position, tokenIndex = position185, tokenIndex185
				}
				{
					position187, tokenIndex187 := position, tokenIndex
					{
						position189, tokenIndex189 := position, tokenIndex
						if buffer[position] != rune('?') {
							goto l190
						}
						position++
						goto l189
					l190:
						position, tokenIndex = position189, tokenIndex189
						if buffer[position] != rune('!') {
							goto l187
						}
						position++
					}
				l189:
					goto l188
				l187:
					position, tokenIndex = position187, tokenIndex187
				}
			l188:
				add(ruleFunLabel, position181)
			}
			return true
		l180:
			position, tokenIndex = position180, tokenIndex180
			return false
		},
		/* 20 Type <- <(PointerType / FunType / GenericType / CapitalLabel / BuiltinType)> */
		func() bool {
			position191, tokenIndex191 := position, tokenIndex
			{
				position192 := position
				{
					position193, tokenIndex193 := position, tokenIndex
					{
						position195 := position
						if buffer[position] != rune('*') {
							goto l194
						}
						position++
						if !_rules[ruleType]() {
							goto l194
						}
						add(rulePointerType, position195)
					}
					goto l193
				l194:
					position, tokenIndex = position193, tokenIndex193
					{
						position197 := position
					l198:
						{
							position199, tokenIndex199 := position, tokenIndex
							if !_rules[ruleTypeExceptFun]() {
								goto l199
							}
							if buffer[position] != rune(',') {
								goto l199
							}
							position++
							{
								position200, tokenIndex200 := position, tokenIndex
								if !_rules[ruleWhitespace]() {
									goto l200
								}
								goto l201
							l200:
								position, tokenIndex = position200, tokenIndex200
							}
						l201:
							goto l198
						l199:
							position, tokenIndex = position199, tokenIndex199
						}
						if !_rules[ruleTypeExceptFun]() {
							goto l196
						}
						if !_rules[ruleWhitespace]() {
							goto l196
						}
						if buffer[position] != rune('-') {
							goto l196
						}
						position++
						if buffer[position] != rune('>') {
							goto l196
						}
						position++
						if !_rules[ruleWhitespace]() {
							goto l196
						}
						if !_rules[ruleTypeExceptFun]() {
							goto l196
						}
						add(ruleFunType, position197)
					}
					goto l193
				l196:
					position, tokenIndex = position193, tokenIndex193
					if !_rules[ruleGenericType]() {
						goto l202
					}
					goto l193
				l202:
					position, tokenIndex = position193, tokenIndex193
					if !_rules[ruleCapitalLabel]() {
						goto l203
					}
					goto l193
				l203:
					position, tokenIndex = position193, tokenIndex193
					if !_rules[ruleBuiltinType]() {
						goto l191
					}
				}
			l193:
				add(ruleType, position192)
			}
			return true
		l191:
			position, tokenIndex = position191, tokenIndex191
			return false
		},
		/* 21 PointerType <- <('*' Type)> */
		nil,
		/* 22 FunType <- <((TypeExceptFun ',' Whitespace?)* TypeExceptFun Whitespace ('-' '>') Whitespace TypeExceptFun)> */
		nil,
		/* 23 GenericType <- <(CapitalLabel '<' (CapitalLabel ',' Whitespace?)* CapitalLabel '>')> */
		func() bool {
			position206, tokenIndex206 := position, tokenIndex
			{
				position207 := position
				if !_rules[ruleCapitalLabel]() {
					goto l206
				}
				if buffer[position] != rune('<') {
					goto l206
				}
				position++
			l208:
				{
					position209, tokenIndex209 := position, tokenIndex
					if !_rules[ruleCapitalLabel]() {
						goto l209
					}
					if buffer[position] != rune(',') {
						goto l209
					}
					position++
					{
						position210, tokenIndex210 := position, tokenIndex
						if !_rules[ruleWhitespace]() {
							goto l210
						}
						goto l211
					l210:
						position, tokenIndex = position210, tokenIndex210
					}
				l211:
					goto l208
				l209:
					position, tokenIndex = position209, tokenIndex209
				}
				if !_rules[ruleCapitalLabel]() {
					goto l206
				}
				if buffer[position] != rune('>') {
					goto l206
				}
				position++
				add(ruleGenericType, position207)
			}
			return true
		l206:
			position, tokenIndex = position206, tokenIndex206
			return false
		},
		/* 24 BuiltinType <- <(BuiltinSlice / ((&('M' | 'm') BuiltinMap) | (&('[') BuiltinArray) | (&('I' | 'R' | 'S' | 'i' | 'r' | 's') BuiltinSimple)))> */
		func() bool {
			position212, tokenIndex212 := position, tokenIndex
			{
				position213 := position
				{
					position214, tokenIndex214 := position, tokenIndex
					{
						position216 := position
						if buffer[position] != rune('[') {
							goto l215
						}
						position++
						if buffer[position] != rune(']') {
							goto l215
						}
						position++
						if !_rules[ruleType]() {
							goto l215
						}
						add(ruleBuiltinSlice, position216)
					}
					goto l214
				l215:
					position, tokenIndex = position214, tokenIndex214
					{
						switch buffer[position] {
						case 'M', 'm':
							{
								position218 := position
								{
									position219, tokenIndex219 := position, tokenIndex
									if buffer[position] != rune('m') {
										goto l220
									}
									position++
									goto l219
								l220:
									position, tokenIndex = position219, tokenIndex219
									if buffer[position] != rune('M') {
										goto l212
									}
									position++
								}
							l219:
								{
									position221, tokenIndex221 := position, tokenIndex
									if buffer[position] != rune('a') {
										goto l222
									}
									position++
									goto l221
								l222:
									position, tokenIndex = position221, tokenIndex221
									if buffer[position] != rune('A') {
										goto l212
									}
									position++
								}
							l221:
								{
									position223, tokenIndex223 := position, tokenIndex
									if buffer[position] != rune('p') {
										goto l224
									}
									position++
									goto l223
								l224:
									position, tokenIndex = position223, tokenIndex223
									if buffer[position] != rune('P') {
										goto l212
									}
									position++
								}
							l223:
								if buffer[position] != rune('[') {
									goto l212
								}
								position++
								if !_rules[ruleType]() {
									goto l212
								}
								if buffer[position] != rune(']') {
									goto l212
								}
								position++
								if !_rules[ruleType]() {
									goto l212
								}
								add(ruleBuiltinMap, position218)
							}
							break
						case '[':
							{
								position225 := position
								if buffer[position] != rune('[') {
									goto l212
								}
								position++
								if !_rules[ruleInteger]() {
									goto l212
								}
								if buffer[position] != rune(']') {
									goto l212
								}
								position++
								if !_rules[ruleType]() {
									goto l212
								}
								add(ruleBuiltinArray, position225)
							}
							break
						default:
							{
								position226 := position
								{
									switch buffer[position] {
									case 'S', 's':
										{
											position228, tokenIndex228 := position, tokenIndex
											if buffer[position] != rune('s') {
												goto l229
											}
											position++
											goto l228
										l229:
											position, tokenIndex = position228, tokenIndex228
											if buffer[position] != rune('S') {
												goto l212
											}
											position++
										}
									l228:
										{
											position230, tokenIndex230 := position, tokenIndex
											if buffer[position] != rune('t') {
												goto l231
											}
											position++
											goto l230
										l231:
											position, tokenIndex = position230, tokenIndex230
											if buffer[position] != rune('T') {
												goto l212
											}
											position++
										}
									l230:
										{
											position232, tokenIndex232 := position, tokenIndex
											if buffer[position] != rune('r') {
												goto l233
											}
											position++
											goto l232
										l233:
											position, tokenIndex = position232, tokenIndex232
											if buffer[position] != rune('R') {
												goto l212
											}
											position++
										}
									l232:
										{
											position234, tokenIndex234 := position, tokenIndex
											if buffer[position] != rune('i') {
												goto l235
											}
											position++
											goto l234
										l235:
											position, tokenIndex = position234, tokenIndex234
											if buffer[position] != rune('I') {
												goto l212
											}
											position++
										}
									l234:
										{
											position236, tokenIndex236 := position, tokenIndex
											if buffer[position] != rune('n') {
												goto l237
											}
											position++
											goto l236
										l237:
											position, tokenIndex = position236, tokenIndex236
											if buffer[position] != rune('N') {
												goto l212
											}
											position++
										}
									l236:
										{
											position238, tokenIndex238 := position, tokenIndex
											if buffer[position] != rune('g') {
												goto l239
											}
											position++
											goto l238
										l239:
											position, tokenIndex = position238, tokenIndex238
											if buffer[position] != rune('G') {
												goto l212
											}
											position++
										}
									l238:
										break
									case 'R', 'r':
										{
											position240, tokenIndex240 := position, tokenIndex
											if buffer[position] != rune('r') {
												goto l241
											}
											position++
											goto l240
										l241:
											position, tokenIndex = position240, tokenIndex240
											if buffer[position] != rune('R') {
												goto l212
											}
											position++
										}
									l240:
										{
											position242, tokenIndex242 := position, tokenIndex
											if buffer[position] != rune('e') {
												goto l243
											}
											position++
											goto l242
										l243:
											position, tokenIndex = position242, tokenIndex242
											if buffer[position] != rune('E') {
												goto l212
											}
											position++
										}
									l242:
										{
											position244, tokenIndex244 := position, tokenIndex
											if buffer[position] != rune('a') {
												goto l245
											}
											position++
											goto l244
										l245:
											position, tokenIndex = position244, tokenIndex244
											if buffer[position] != rune('A') {
												goto l212
											}
											position++
										}
									l244:
										{
											position246, tokenIndex246 := position, tokenIndex
											if buffer[position] != rune('l') {
												goto l247
											}
											position++
											goto l246
										l247:
											position, tokenIndex = position246, tokenIndex246
											if buffer[position] != rune('L') {
												goto l212
											}
											position++
										}
									l246:
										break
									default:
										{
											position248, tokenIndex248 := position, tokenIndex
											if buffer[position] != rune('i') {
												goto l249
											}
											position++
											goto l248
										l249:
											position, tokenIndex = position248, tokenIndex248
											if buffer[position] != rune('I') {
												goto l212
											}
											position++
										}
									l248:
										{
											position250, tokenIndex250 := position, tokenIndex
											if buffer[position] != rune('n') {
												goto l251
											}
											position++
											goto l250
										l251:
											position, tokenIndex = position250, tokenIndex250
											if buffer[position] != rune('N') {
												goto l212
											}
											position++
										}
									l250:
										{
											position252, tokenIndex252 := position, tokenIndex
											if buffer[position] != rune('t') {
												goto l253
											}
											position++
											goto l252
										l253:
											position, tokenIndex = position252, tokenIndex252
											if buffer[position] != rune('T') {
												goto l212
											}
											position++
										}
									l252:
										break
									}
								}

								add(ruleBuiltinSimple, position226)
							}
							break
						}
					}

				}
			l214:
				add(ruleBuiltinType, position213)
			}
			return true
		l212:
			position, tokenIndex = position212, tokenIndex212
			return false
		},
		/* 25 BuiltinSimple <- <((&('S' | 's') (('s' / 'S') ('t' / 'T') ('r' / 'R') ('i' / 'I') ('n' / 'N') ('g' / 'G'))) | (&('R' | 'r') (('r' / 'R') ('e' / 'E') ('a' / 'A') ('l' / 'L'))) | (&('I' | 'i') (('i' / 'I') ('n' / 'N') ('t' / 'T'))))> */
		nil,
		/* 26 BuiltinSlice <- <('[' ']' Type)> */
		nil,
		/* 27 BuiltinArray <- <('[' Integer ']' Type)> */
		nil,
		/* 28 BuiltinMap <- <(('m' / 'M') ('a' / 'A') ('p' / 'P') '[' Type ']' Type)> */
		nil,
		/* 29 TypeExceptFun <- <(GenericType / BuiltinType / CapitalLabel)> */
		func() bool {
			position258, tokenIndex258 := position, tokenIndex
			{
				position259 := position
				{
					position260, tokenIndex260 := position, tokenIndex
					if !_rules[ruleGenericType]() {
						goto l261
					}
					goto l260
				l261:
					position, tokenIndex = position260, tokenIndex260
					if !_rules[ruleBuiltinType]() {
						goto l262
					}
					goto l260
				l262:
					position, tokenIndex = position260, tokenIndex260
					if !_rules[ruleCapitalLabel]() {
						goto l258
					}
				}
			l260:
				add(ruleTypeExceptFun, position259)
			}
			return true
		l258:
			position, tokenIndex = position258, tokenIndex258
			return false
		},
		/* 30 Code <- <((Line Newline)+ Dedent)> */
		func() bool {
			position263, tokenIndex263 := position, tokenIndex
			{
				position264 := position
				{
					position267 := position
					{
						position268, tokenIndex268 := position, tokenIndex
						{
							position270 := position
							if !_rules[ruleExpression]() {
								goto l269
							}
							if buffer[position] != rune('[') {
								goto l269
							}
							position++
							if !_rules[ruleExpression]() {
								goto l269
							}
							if buffer[position] != rune(']') {
								goto l269
							}
							position++
							if !_rules[ruleWhitespace]() {
								goto l269
							}
							if buffer[position] != rune('=') {
								goto l269
							}
							position++
							if !_rules[ruleWhitespace]() {
								goto l269
							}
							if !_rules[ruleExpression]() {
								goto l269
							}
							add(ruleIndexAssignment, position270)
						}
						goto l268
					l269:
						position, tokenIndex = position268, tokenIndex268
						{
							position272 := position
							if !_rules[ruleLowerLabel]() {
								goto l271
							}
							if !_rules[ruleWhitespace]() {
								goto l271
							}
							if buffer[position] != rune('=') {
								goto l271
							}
							position++
							if !_rules[ruleWhitespace]() {
								goto l271
							}
							if !_rules[ruleExpression]() {
								goto l271
							}
							add(ruleAssignment, position272)
						}
						goto l268
					l271:
						position, tokenIndex = position268, tokenIndex268
						if !_rules[ruleBinaryOperation]() {
							goto l273
						}
						goto l268
					l273:
						position, tokenIndex = position268, tokenIndex268
						if !_rules[ruleCall]() {
							goto l274
						}
						goto l268
					l274:
						position, tokenIndex = position268, tokenIndex268
						{
							switch buffer[position] {
							case 'O', 'o':
								{
									position276 := position
									{
										position277, tokenIndex277 := position, tokenIndex
										if buffer[position] != rune('o') {
											goto l278
										}
										position++
										goto l277
									l278:
										position, tokenIndex = position277, tokenIndex277
										if buffer[position] != rune('O') {
											goto l263
										}
										position++
									}
								l277:
									{
										position279, tokenIndex279 := position, tokenIndex
										if buffer[position] != rune('n') {
											goto l280
										}
										position++
										goto l279
									l280:
										position, tokenIndex = position279, tokenIndex279
										if buffer[position] != rune('N') {
											goto l263
										}
										position++
									}
								l279:
									if !_rules[ruleWhitespace]() {
										goto l263
									}
									if !_rules[ruleFunLabel]() {
										goto l263
									}
									if buffer[position] != rune(':') {
										goto l263
									}
									position++
									if !_rules[ruleNewline]() {
										goto l263
									}
									if !_rules[ruleIndent]() {
										goto l263
									}
									if !_rules[ruleCode]() {
										goto l263
									}
									add(ruleOn, position276)
								}
								break
							case 'F', 'f':
								{
									position281 := position
									{
										position282, tokenIndex282 := position, tokenIndex
										{
											position284 := position
											{
												position285, tokenIndex285 := position, tokenIndex
												if buffer[position] != rune('f') {
													goto l286
												}
												position++
												goto l285
											l286:
												position, tokenIndex = position285, tokenIndex285
												if buffer[position] != rune('F') {
													goto l283
												}
												position++
											}
										l285:
											{
												position287, tokenIndex287 := position, tokenIndex
												if buffer[position] != rune('o') {
													goto l288
												}
												position++
												goto l287
											l288:
												position, tokenIndex = position287, tokenIndex287
												if buffer[position] != rune('O') {
													goto l283
												}
												position++
											}
										l287:
											{
												position289, tokenIndex289 := position, tokenIndex
												if buffer[position] != rune('r') {
													goto l290
												}
												position++
												goto l289
											l290:
												position, tokenIndex = position289, tokenIndex289
												if buffer[position] != rune('R') {
													goto l283
												}
												position++
											}
										l289:
											if !_rules[ruleWhitespace]() {
												goto l283
											}
										l291:
											{
												position292, tokenIndex292 := position, tokenIndex
												if !_rules[ruleLowerLabel]() {
													goto l292
												}
												if buffer[position] != rune(',') {
													goto l292
												}
												position++
												{
													position293, tokenIndex293 := position, tokenIndex
													if !_rules[ruleWhitespace]() {
														goto l293
													}
													goto l294
												l293:
													position, tokenIndex = position293, tokenIndex293
												}
											l294:
												goto l291
											l292:
												position, tokenIndex = position292, tokenIndex292
											}
											if !_rules[ruleLowerLabel]() {
												goto l283
											}
											if !_rules[ruleWhitespace]() {
												goto l283
											}
											if buffer[position] != rune('i') {
												goto l283
											}
											position++
											if buffer[position] != rune('n') {
												goto l283
											}
											position++
											if !_rules[ruleWhitespace]() {
												goto l283
											}
											if !_rules[ruleExpression]() {
												goto l283
											}
											if buffer[position] != rune(':') {
												goto l283
											}
											position++
											if !_rules[ruleNewline]() {
												goto l283
											}
											if !_rules[ruleIndent]() {
												goto l283
											}
											if !_rules[ruleCode]() {
												goto l283
											}
											add(ruleForIn, position284)
										}
										goto l282
									l283:
										position, tokenIndex = position282, tokenIndex282
										{
											position295 := position
											{
												position296, tokenIndex296 := position, tokenIndex
												if buffer[position] != rune('f') {
													goto l297
												}
												position++
												goto l296
											l297:
												position, tokenIndex = position296, tokenIndex296
												if buffer[position] != rune('F') {
													goto l263
												}
												position++
											}
										l296:
											{
												position298, tokenIndex298 := position, tokenIndex
												if buffer[position] != rune('o') {
													goto l299
												}
												position++
												goto l298
											l299:
												position, tokenIndex = position298, tokenIndex298
												if buffer[position] != rune('O') {
													goto l263
												}
												position++
											}
										l298:
											{
												position300, tokenIndex300 := position, tokenIndex
												if buffer[position] != rune('r') {
													goto l301
												}
												position++
												goto l300
											l301:
												position, tokenIndex = position300, tokenIndex300
												if buffer[position] != rune('R') {
													goto l263
												}
												position++
											}
										l300:
											if !_rules[ruleWhitespace]() {
												goto l263
											}
											if !_rules[ruleLowerLabel]() {
												goto l263
											}
											if !_rules[ruleWhitespace]() {
												goto l263
											}
											if buffer[position] != rune('i') {
												goto l263
											}
											position++
											if buffer[position] != rune('n') {
												goto l263
											}
											position++
											if !_rules[ruleWhitespace]() {
												goto l263
											}
											{
												position302 := position
												if !_rules[ruleInteger]() {
													goto l263
												}
												{
													position303 := position
													{
														position304, tokenIndex304 := position, tokenIndex
														if buffer[position] != rune('.') {
															goto l305
														}
														position++
														if buffer[position] != rune('.') {
															goto l305
														}
														position++
														if buffer[position] != rune('.') {
															goto l305
														}
														position++
														goto l304
													l305:
														position, tokenIndex = position304, tokenIndex304
														if buffer[position] != rune('.') {
															goto l263
														}
														position++
														if buffer[position] != rune('.') {
															goto l263
														}
														position++
													}
												l304:
													add(ruleRangeOperator, position303)
												}
												if !_rules[ruleInteger]() {
													goto l263
												}
												add(ruleRange, position302)
											}
											if buffer[position] != rune(':') {
												goto l263
											}
											position++
											if !_rules[ruleNewline]() {
												goto l263
											}
											if !_rules[ruleIndent]() {
												goto l263
											}
											if !_rules[ruleCode]() {
												goto l263
											}
											add(ruleForLoop, position295)
										}
									}
								l282:
									add(ruleFor, position281)
								}
								break
							case '+', '-':
								if !_rules[ruleUnaryOperation]() {
									goto l263
								}
								break
							default:
								{
									position306 := position
									{
										switch buffer[position] {
										case 'E', 'e':
											{
												position308 := position
												{
													position309, tokenIndex309 := position, tokenIndex
													if buffer[position] != rune('e') {
														goto l310
													}
													position++
													goto l309
												l310:
													position, tokenIndex = position309, tokenIndex309
													if buffer[position] != rune('E') {
														goto l263
													}
													position++
												}
											l309:
												{
													position311, tokenIndex311 := position, tokenIndex
													if buffer[position] != rune('s') {
														goto l312
													}
													position++
													goto l311
												l312:
													position, tokenIndex = position311, tokenIndex311
													if buffer[position] != rune('S') {
														goto l263
													}
													position++
												}
											l311:
												{
													position313, tokenIndex313 := position, tokenIndex
													if buffer[position] != rune('c') {
														goto l314
													}
													position++
													goto l313
												l314:
													position, tokenIndex = position313, tokenIndex313
													if buffer[position] != rune('C') {
														goto l263
													}
													position++
												}
											l313:
												{
													position315, tokenIndex315 := position, tokenIndex
													if buffer[position] != rune('a') {
														goto l316
													}
													position++
													goto l315
												l316:
													position, tokenIndex = position315, tokenIndex315
													if buffer[position] != rune('A') {
														goto l263
													}
													position++
												}
											l315:
												{
													position317, tokenIndex317 := position, tokenIndex
													if buffer[position] != rune('l') {
														goto l318
													}
													position++
													goto l317
												l318:
													position, tokenIndex = position317, tokenIndex317
													if buffer[position] != rune('L') {
														goto l263
													}
													position++
												}
											l317:
												{
													position319, tokenIndex319 := position, tokenIndex
													if buffer[position] != rune('a') {
														goto l320
													}
													position++
													goto l319
												l320:
													position, tokenIndex = position319, tokenIndex319
													if buffer[position] != rune('A') {
														goto l263
													}
													position++
												}
											l319:
												{
													position321, tokenIndex321 := position, tokenIndex
													if buffer[position] != rune('t') {
														goto l322
													}
													position++
													goto l321
												l322:
													position, tokenIndex = position321, tokenIndex321
													if buffer[position] != rune('T') {
														goto l263
													}
													position++
												}
											l321:
												{
													position323, tokenIndex323 := position, tokenIndex
													if buffer[position] != rune('e') {
														goto l324
													}
													position++
													goto l323
												l324:
													position, tokenIndex = position323, tokenIndex323
													if buffer[position] != rune('E') {
														goto l263
													}
													position++
												}
											l323:
												if !_rules[ruleWhitespace]() {
													goto l263
												}
											l325:
												{
													position326, tokenIndex326 := position, tokenIndex
													if !_rules[ruleFunLabel]() {
														goto l326
													}
													if buffer[position] != rune(',') {
														goto l326
													}
													position++
													{
														position327, tokenIndex327 := position, tokenIndex
														if !_rules[ruleWhitespace]() {
															goto l327
														}
														goto l328
													l327:
														position, tokenIndex = position327, tokenIndex327
													}
												l328:
													goto l325
												l326:
													position, tokenIndex = position326, tokenIndex326
												}
												if !_rules[ruleFunLabel]() {
													goto l263
												}
												add(ruleEscalator, position308)
											}
											break
										case '!':
											{
												position329 := position
												if buffer[position] != rune('!') {
													goto l263
												}
												position++
												if buffer[position] != rune('!') {
													goto l263
												}
												position++
												{
													position330, tokenIndex330 := position, tokenIndex
													if !_rules[ruleWhitespace]() {
														goto l330
													}
													goto l331
												l330:
													position, tokenIndex = position330, tokenIndex330
												}
											l331:
												if !_rules[ruleExpression]() {
													goto l263
												}
												add(ruleReturnError, position329)
											}
											break
										default:
											{
												position332 := position
												{
													position333, tokenIndex333 := position, tokenIndex
													if buffer[position] != rune('r') {
														goto l334
													}
													position++
													goto l333
												l334:
													position, tokenIndex = position333, tokenIndex333
													if buffer[position] != rune('R') {
														goto l263
													}
													position++
												}
											l333:
												{
													position335, tokenIndex335 := position, tokenIndex
													if buffer[position] != rune('e') {
														goto l336
													}
													position++
													goto l335
												l336:
													position, tokenIndex = position335, tokenIndex335
													if buffer[position] != rune('E') {
														goto l263
													}
													position++
												}
											l335:
												{
													position337, tokenIndex337 := position, tokenIndex
													if buffer[position] != rune('t') {
														goto l338
													}
													position++
													goto l337
												l338:
													position, tokenIndex = position337, tokenIndex337
													if buffer[position] != rune('T') {
														goto l263
													}
													position++
												}
											l337:
												{
													position339, tokenIndex339 := position, tokenIndex
													if buffer[position] != rune('u') {
														goto l340
													}
													position++
													goto l339
												l340:
													position, tokenIndex = position339, tokenIndex339
													if buffer[position] != rune('U') {
														goto l263
													}
													position++
												}
											l339:
												{
													position341, tokenIndex341 := position, tokenIndex
													if buffer[position] != rune('r') {
														goto l342
													}
													position++
													goto l341
												l342:
													position, tokenIndex = position341, tokenIndex341
													if buffer[position] != rune('R') {
														goto l263
													}
													position++
												}
											l341:
												{
													position343, tokenIndex343 := position, tokenIndex
													if buffer[position] != rune('n') {
														goto l344
													}
													position++
													goto l343
												l344:
													position, tokenIndex = position343, tokenIndex343
													if buffer[position] != rune('N') {
														goto l263
													}
													position++
												}
											l343:
												{
													position345, tokenIndex345 := position, tokenIndex
													if !_rules[ruleWhitespace]() {
														goto l345
													}
													goto l346
												l345:
													position, tokenIndex = position345, tokenIndex345
												}
											l346:
												if !_rules[ruleExpression]() {
													goto l263
												}
												add(ruleReturnValue, position332)
											}
											break
										}
									}

									add(ruleReturn, position306)
								}
								break
							}
						}

					}
				l268:
					add(ruleLine, position267)
				}
				if !_rules[ruleNewline]() {
					goto l263
				}
			l265:
				{
					position266, tokenIndex266 := position, tokenIndex
					{
						position347 := position
						{
							position348, tokenIndex348 := position, tokenIndex
							{
								position350 := position
								if !_rules[ruleExpression]() {
									goto l349
								}
								if buffer[position] != rune('[') {
									goto l349
								}
								position++
								if !_rules[ruleExpression]() {
									goto l349
								}
								if buffer[position] != rune(']') {
									goto l349
								}
								position++
								if !_rules[ruleWhitespace]() {
									goto l349
								}
								if buffer[position] != rune('=') {
									goto l349
								}
								position++
								if !_rules[ruleWhitespace]() {
									goto l349
								}
								if !_rules[ruleExpression]() {
									goto l349
								}
								add(ruleIndexAssignment, position350)
							}
							goto l348
						l349:
							position, tokenIndex = position348, tokenIndex348
							{
								position352 := position
								if !_rules[ruleLowerLabel]() {
									goto l351
								}
								if !_rules[ruleWhitespace]() {
									goto l351
								}
								if buffer[position] != rune('=') {
									goto l351
								}
								position++
								if !_rules[ruleWhitespace]() {
									goto l351
								}
								if !_rules[ruleExpression]() {
									goto l351
								}
								add(ruleAssignment, position352)
							}
							goto l348
						l351:
							position, tokenIndex = position348, tokenIndex348
							if !_rules[ruleBinaryOperation]() {
								goto l353
							}
							goto l348
						l353:
							position, tokenIndex = position348, tokenIndex348
							if !_rules[ruleCall]() {
								goto l354
							}
							goto l348
						l354:
							position, tokenIndex = position348, tokenIndex348
							{
								switch buffer[position] {
								case 'O', 'o':
									{
										position356 := position
										{
											position357, tokenIndex357 := position, tokenIndex
											if buffer[position] != rune('o') {
												goto l358
											}
											position++
											goto l357
										l358:
											position, tokenIndex = position357, tokenIndex357
											if buffer[position] != rune('O') {
												goto l266
											}
											position++
										}
									l357:
										{
											position359, tokenIndex359 := position, tokenIndex
											if buffer[position] != rune('n') {
												goto l360
											}
											position++
											goto l359
										l360:
											position, tokenIndex = position359, tokenIndex359
											if buffer[position] != rune('N') {
												goto l266
											}
											position++
										}
									l359:
										if !_rules[ruleWhitespace]() {
											goto l266
										}
										if !_rules[ruleFunLabel]() {
											goto l266
										}
										if buffer[position] != rune(':') {
											goto l266
										}
										position++
										if !_rules[ruleNewline]() {
											goto l266
										}
										if !_rules[ruleIndent]() {
											goto l266
										}
										if !_rules[ruleCode]() {
											goto l266
										}
										add(ruleOn, position356)
									}
									break
								case 'F', 'f':
									{
										position361 := position
										{
											position362, tokenIndex362 := position, tokenIndex
											{
												position364 := position
												{
													position365, tokenIndex365 := position, tokenIndex
													if buffer[position] != rune('f') {
														goto l366
													}
													position++
													goto l365
												l366:
													position, tokenIndex = position365, tokenIndex365
													if buffer[position] != rune('F') {
														goto l363
													}
													position++
												}
											l365:
												{
													position367, tokenIndex367 := position, tokenIndex
													if buffer[position] != rune('o') {
														goto l368
													}
													position++
													goto l367
												l368:
													position, tokenIndex = position367, tokenIndex367
													if buffer[position] != rune('O') {
														goto l363
													}
													position++
												}
											l367:
												{
													position369, tokenIndex369 := position, tokenIndex
													if buffer[position] != rune('r') {
														goto l370
													}
													position++
													goto l369
												l370:
													position, tokenIndex = position369, tokenIndex369
													if buffer[position] != rune('R') {
														goto l363
													}
													position++
												}
											l369:
												if !_rules[ruleWhitespace]() {
													goto l363
												}
											l371:
												{
													position372, tokenIndex372 := position, tokenIndex
													if !_rules[ruleLowerLabel]() {
														goto l372
													}
													if buffer[position] != rune(',') {
														goto l372
													}
													position++
													{
														position373, tokenIndex373 := position, tokenIndex
														if !_rules[ruleWhitespace]() {
															goto l373
														}
														goto l374
													l373:
														position, tokenIndex = position373, tokenIndex373
													}
												l374:
													goto l371
												l372:
													position, tokenIndex = position372, tokenIndex372
												}
												if !_rules[ruleLowerLabel]() {
													goto l363
												}
												if !_rules[ruleWhitespace]() {
													goto l363
												}
												if buffer[position] != rune('i') {
													goto l363
												}
												position++
												if buffer[position] != rune('n') {
													goto l363
												}
												position++
												if !_rules[ruleWhitespace]() {
													goto l363
												}
												if !_rules[ruleExpression]() {
													goto l363
												}
												if buffer[position] != rune(':') {
													goto l363
												}
												position++
												if !_rules[ruleNewline]() {
													goto l363
												}
												if !_rules[ruleIndent]() {
													goto l363
												}
												if !_rules[ruleCode]() {
													goto l363
												}
												add(ruleForIn, position364)
											}
											goto l362
										l363:
											position, tokenIndex = position362, tokenIndex362
											{
												position375 := position
												{
													position376, tokenIndex376 := position, tokenIndex
													if buffer[position] != rune('f') {
														goto l377
													}
													position++
													goto l376
												l377:
													position, tokenIndex = position376, tokenIndex376
													if buffer[position] != rune('F') {
														goto l266
													}
													position++
												}
											l376:
												{
													position378, tokenIndex378 := position, tokenIndex
													if buffer[position] != rune('o') {
														goto l379
													}
													position++
													goto l378
												l379:
													position, tokenIndex = position378, tokenIndex378
													if buffer[position] != rune('O') {
														goto l266
													}
													position++
												}
											l378:
												{
													position380, tokenIndex380 := position, tokenIndex
													if buffer[position] != rune('r') {
														goto l381
													}
													position++
													goto l380
												l381:
													position, tokenIndex = position380, tokenIndex380
													if buffer[position] != rune('R') {
														goto l266
													}
													position++
												}
											l380:
												if !_rules[ruleWhitespace]() {
													goto l266
												}
												if !_rules[ruleLowerLabel]() {
													goto l266
												}
												if !_rules[ruleWhitespace]() {
													goto l266
												}
												if buffer[position] != rune('i') {
													goto l266
												}
												position++
												if buffer[position] != rune('n') {
													goto l266
												}
												position++
												if !_rules[ruleWhitespace]() {
													goto l266
												}
												{
													position382 := position
													if !_rules[ruleInteger]() {
														goto l266
													}
													{
														position383 := position
														{
															position384, tokenIndex384 := position, tokenIndex
															if buffer[position] != rune('.') {
																goto l385
															}
															position++
															if buffer[position] != rune('.') {
																goto l385
															}
															position++
															if buffer[position] != rune('.') {
																goto l385
															}
															position++
															goto l384
														l385:
															position, tokenIndex = position384, tokenIndex384
															if buffer[position] != rune('.') {
																goto l266
															}
															position++
															if buffer[position] != rune('.') {
																goto l266
															}
															position++
														}
													l384:
														add(ruleRangeOperator, position383)
													}
													if !_rules[ruleInteger]() {
														goto l266
													}
													add(ruleRange, position382)
												}
												if buffer[position] != rune(':') {
													goto l266
												}
												position++
												if !_rules[ruleNewline]() {
													goto l266
												}
												if !_rules[ruleIndent]() {
													goto l266
												}
												if !_rules[ruleCode]() {
													goto l266
												}
												add(ruleForLoop, position375)
											}
										}
									l362:
										add(ruleFor, position361)
									}
									break
								case '+', '-':
									if !_rules[ruleUnaryOperation]() {
										goto l266
									}
									break
								default:
									{
										position386 := position
										{
											switch buffer[position] {
											case 'E', 'e':
												{
													position388 := position
													{
														position389, tokenIndex389 := position, tokenIndex
														if buffer[position] != rune('e') {
															goto l390
														}
														position++
														goto l389
													l390:
														position, tokenIndex = position389, tokenIndex389
														if buffer[position] != rune('E') {
															goto l266
														}
														position++
													}
												l389:
													{
														position391, tokenIndex391 := position, tokenIndex
														if buffer[position] != rune('s') {
															goto l392
														}
														position++
														goto l391
													l392:
														position, tokenIndex = position391, tokenIndex391
														if buffer[position] != rune('S') {
															goto l266
														}
														position++
													}
												l391:
													{
														position393, tokenIndex393 := position, tokenIndex
														if buffer[position] != rune('c') {
															goto l394
														}
														position++
														goto l393
													l394:
														position, tokenIndex = position393, tokenIndex393
														if buffer[position] != rune('C') {
															goto l266
														}
														position++
													}
												l393:
													{
														position395, tokenIndex395 := position, tokenIndex
														if buffer[position] != rune('a') {
															goto l396
														}
														position++
														goto l395
													l396:
														position, tokenIndex = position395, tokenIndex395
														if buffer[position] != rune('A') {
															goto l266
														}
														position++
													}
												l395:
													{
														position397, tokenIndex397 := position, tokenIndex
														if buffer[position] != rune('l') {
															goto l398
														}
														position++
														goto l397
													l398:
														position, tokenIndex = position397, tokenIndex397
														if buffer[position] != rune('L') {
															goto l266
														}
														position++
													}
												l397:
													{
														position399, tokenIndex399 := position, tokenIndex
														if buffer[position] != rune('a') {
															goto l400
														}
														position++
														goto l399
													l400:
														position, tokenIndex = position399, tokenIndex399
														if buffer[position] != rune('A') {
															goto l266
														}
														position++
													}
												l399:
													{
														position401, tokenIndex401 := position, tokenIndex
														if buffer[position] != rune('t') {
															goto l402
														}
														position++
														goto l401
													l402:
														position, tokenIndex = position401, tokenIndex401
														if buffer[position] != rune('T') {
															goto l266
														}
														position++
													}
												l401:
													{
														position403, tokenIndex403 := position, tokenIndex
														if buffer[position] != rune('e') {
															goto l404
														}
														position++
														goto l403
													l404:
														position, tokenIndex = position403, tokenIndex403
														if buffer[position] != rune('E') {
															goto l266
														}
														position++
													}
												l403:
													if !_rules[ruleWhitespace]() {
														goto l266
													}
												l405:
													{
														position406, tokenIndex406 := position, tokenIndex
														if !_rules[ruleFunLabel]() {
															goto l406
														}
														if buffer[position] != rune(',') {
															goto l406
														}
														position++
														{
															position407, tokenIndex407 := position, tokenIndex
															if !_rules[ruleWhitespace]() {
																goto l407
															}
															goto l408
														l407:
															position, tokenIndex = position407, tokenIndex407
														}
													l408:
														goto l405
													l406:
														position, tokenIndex = position406, tokenIndex406
													}
													if !_rules[ruleFunLabel]() {
														goto l266
													}
													add(ruleEscalator, position388)
												}
												break
											case '!':
												{
													position409 := position
													if buffer[position] != rune('!') {
														goto l266
													}
													position++
													if buffer[position] != rune('!') {
														goto l266
													}
													position++
													{
														position410, tokenIndex410 := position, tokenIndex
														if !_rules[ruleWhitespace]() {
															goto l410
														}
														goto l411
													l410:
														position, tokenIndex = position410, tokenIndex410
													}
												l411:
													if !_rules[ruleExpression]() {
														goto l266
													}
													add(ruleReturnError, position409)
												}
												break
											default:
												{
													position412 := position
													{
														position413, tokenIndex413 := position, tokenIndex
														if buffer[position] != rune('r') {
															goto l414
														}
														position++
														goto l413
													l414:
														position, tokenIndex = position413, tokenIndex413
														if buffer[position] != rune('R') {
															goto l266
														}
														position++
													}
												l413:
													{
														position415, tokenIndex415 := position, tokenIndex
														if buffer[position] != rune('e') {
															goto l416
														}
														position++
														goto l415
													l416:
														position, tokenIndex = position415, tokenIndex415
														if buffer[position] != rune('E') {
															goto l266
														}
														position++
													}
												l415:
													{
														position417, tokenIndex417 := position, tokenIndex
														if buffer[position] != rune('t') {
															goto l418
														}
														position++
														goto l417
													l418:
														position, tokenIndex = position417, tokenIndex417
														if buffer[position] != rune('T') {
															goto l266
														}
														position++
													}
												l417:
													{
														position419, tokenIndex419 := position, tokenIndex
														if buffer[position] != rune('u') {
															goto l420
														}
														position++
														goto l419
													l420:
														position, tokenIndex = position419, tokenIndex419
														if buffer[position] != rune('U') {
															goto l266
														}
														position++
													}
												l419:
													{
														position421, tokenIndex421 := position, tokenIndex
														if buffer[position] != rune('r') {
															goto l422
														}
														position++
														goto l421
													l422:
														position, tokenIndex = position421, tokenIndex421
														if buffer[position] != rune('R') {
															goto l266
														}
														position++
													}
												l421:
													{
														position423, tokenIndex423 := position, tokenIndex
														if buffer[position] != rune('n') {
															goto l424
														}
														position++
														goto l423
													l424:
														position, tokenIndex = position423, tokenIndex423
														if buffer[position] != rune('N') {
															goto l266
														}
														position++
													}
												l423:
													{
														position425, tokenIndex425 := position, tokenIndex
														if !_rules[ruleWhitespace]() {
															goto l425
														}
														goto l426
													l425:
														position, tokenIndex = position425, tokenIndex425
													}
												l426:
													if !_rules[ruleExpression]() {
														goto l266
													}
													add(ruleReturnValue, position412)
												}
												break
											}
										}

										add(ruleReturn, position386)
									}
									break
								}
							}

						}
					l348:
						add(ruleLine, position347)
					}
					if !_rules[ruleNewline]() {
						goto l266
					}
					goto l265
				l266:
					position, tokenIndex = position266, tokenIndex266
				}
				if !_rules[ruleDedent]() {
					goto l263
				}
				add(ruleCode, position264)
			}
			return true
		l263:
			position, tokenIndex = position263, tokenIndex263
			return false
		},
		/* 31 Indent <- <('@' '@' ('i' / 'I') ('n' / 'N') ('d' / 'D') ('e' / 'E') ('n' / 'N') ('t' / 'T') '@' '@')> */
		func() bool {
			position427, tokenIndex427 := position, tokenIndex
			{
				position428 := position
				if buffer[position] != rune('@') {
					goto l427
				}
				position++
				if buffer[position] != rune('@') {
					goto l427
				}
				position++
				{
					position429, tokenIndex429 := position, tokenIndex
					if buffer[position] != rune('i') {
						goto l430
					}
					position++
					goto l429
				l430:
					position, tokenIndex = position429, tokenIndex429
					if buffer[position] != rune('I') {
						goto l427
					}
					position++
				}
			l429:
				{
					position431, tokenIndex431 := position, tokenIndex
					if buffer[position] != rune('n') {
						goto l432
					}
					position++
					goto l431
				l432:
					position, tokenIndex = position431, tokenIndex431
					if buffer[position] != rune('N') {
						goto l427
					}
					position++
				}
			l431:
				{
					position433, tokenIndex433 := position, tokenIndex
					if buffer[position] != rune('d') {
						goto l434
					}
					position++
					goto l433
				l434:
					position, tokenIndex = position433, tokenIndex433
					if buffer[position] != rune('D') {
						goto l427
					}
					position++
				}
			l433:
				{
					position435, tokenIndex435 := position, tokenIndex
					if buffer[position] != rune('e') {
						goto l436
					}
					position++
					goto l435
				l436:
					position, tokenIndex = position435, tokenIndex435
					if buffer[position] != rune('E') {
						goto l427
					}
					position++
				}
			l435:
				{
					position437, tokenIndex437 := position, tokenIndex
					if buffer[position] != rune('n') {
						goto l438
					}
					position++
					goto l437
				l438:
					position, tokenIndex = position437, tokenIndex437
					if buffer[position] != rune('N') {
						goto l427
					}
					position++
				}
			l437:
				{
					position439, tokenIndex439 := position, tokenIndex
					if buffer[position] != rune('t') {
						goto l440
					}
					position++
					goto l439
				l440:
					position, tokenIndex = position439, tokenIndex439
					if buffer[position] != rune('T') {
						goto l427
					}
					position++
				}
			l439:
				if buffer[position] != rune('@') {
					goto l427
				}
				position++
				if buffer[position] != rune('@') {
					goto l427
				}
				position++
				add(ruleIndent, position428)
			}
			return true
		l427:
			position, tokenIndex = position427, tokenIndex427
			return false
		},
		/* 32 Dedent <- <('@' '@' ('d' / 'D') ('e' / 'E') ('d' / 'D') ('e' / 'E') ('n' / 'N') ('t' / 'T') '@' '@')> */
		func() bool {
			position441, tokenIndex441 := position, tokenIndex
			{
				position442 := position
				if buffer[position] != rune('@') {
					goto l441
				}
				position++
				if buffer[position] != rune('@') {
					goto l441
				}
				position++
				{
					position443, tokenIndex443 := position, tokenIndex
					if buffer[position] != rune('d') {
						goto l444
					}
					position++
					goto l443
				l444:
					position, tokenIndex = position443, tokenIndex443
					if buffer[position] != rune('D') {
						goto l441
					}
					position++
				}
			l443:
				{
					position445, tokenIndex445 := position, tokenIndex
					if buffer[position] != rune('e') {
						goto l446
					}
					position++
					goto l445
				l446:
					position, tokenIndex = position445, tokenIndex445
					if buffer[position] != rune('E') {
						goto l441
					}
					position++
				}
			l445:
				{
					position447, tokenIndex447 := position, tokenIndex
					if buffer[position] != rune('d') {
						goto l448
					}
					position++
					goto l447
				l448:
					position, tokenIndex = position447, tokenIndex447
					if buffer[position] != rune('D') {
						goto l441
					}
					position++
				}
			l447:
				{
					position449, tokenIndex449 := position, tokenIndex
					if buffer[position] != rune('e') {
						goto l450
					}
					position++
					goto l449
				l450:
					position, tokenIndex = position449, tokenIndex449
					if buffer[position] != rune('E') {
						goto l441
					}
					position++
				}
			l449:
				{
					position451, tokenIndex451 := position, tokenIndex
					if buffer[position] != rune('n') {
						goto l452
					}
					position++
					goto l451
				l452:
					position, tokenIndex = position451, tokenIndex451
					if buffer[position] != rune('N') {
						goto l441
					}
					position++
				}
			l451:
				{
					position453, tokenIndex453 := position, tokenIndex
					if buffer[position] != rune('t') {
						goto l454
					}
					position++
					goto l453
				l454:
					position, tokenIndex = position453, tokenIndex453
					if buffer[position] != rune('T') {
						goto l441
					}
					position++
				}
			l453:
				if buffer[position] != rune('@') {
					goto l441
				}
				position++
				if buffer[position] != rune('@') {
					goto l441
				}
				position++
				add(ruleDedent, position442)
			}
			return true
		l441:
			position, tokenIndex = position441, tokenIndex441
			return false
		},
		/* 33 Line <- <(IndexAssignment / Assignment / BinaryOperation / Call / ((&('O' | 'o') On) | (&('F' | 'f') For) | (&('+' | '-') UnaryOperation) | (&('!' | 'E' | 'R' | 'e' | 'r') Return)))> */
		nil,
		/* 34 IndexAssignment <- <(Expression '[' Expression ']' Whitespace '=' Whitespace Expression)> */
		nil,
		/* 35 Assignment <- <(LowerLabel Whitespace '=' Whitespace Expression)> */
		nil,
		/* 36 Expression <- <(BinaryOperation / UnaryOperation / Call / Simple)> */
		func() bool {
			position458, tokenIndex458 := position, tokenIndex
			{
				position459 := position
				{
					position460, tokenIndex460 := position, tokenIndex
					if !_rules[ruleBinaryOperation]() {
						goto l461
					}
					goto l460
				l461:
					position, tokenIndex = position460, tokenIndex460
					if !_rules[ruleUnaryOperation]() {
						goto l462
					}
					goto l460
				l462:
					position, tokenIndex = position460, tokenIndex460
					if !_rules[ruleCall]() {
						goto l463
					}
					goto l460
				l463:
					position, tokenIndex = position460, tokenIndex460
					if !_rules[ruleSimple]() {
						goto l458
					}
				}
			l460:
				add(ruleExpression, position459)
			}
			return true
		l458:
			position, tokenIndex = position458, tokenIndex458
			return false
		},
		/* 37 Simple <- <(Label / ((&('$') Error) | (&('"') String) | (&('F' | 'N' | 'T' | 'f' | 'n' | 't') Constant) | (&('[') List) | (&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') Number)))> */
		func() bool {
			position464, tokenIndex464 := position, tokenIndex
			{
				position465 := position
				{
					position466, tokenIndex466 := position, tokenIndex
					if !_rules[ruleLabel]() {
						goto l467
					}
					goto l466
				l467:
					position, tokenIndex = position466, tokenIndex466
					{
						switch buffer[position] {
						case '$':
							{
								position469 := position
								if buffer[position] != rune('$') {
									goto l464
								}
								position++
								if !_rules[ruleLabel]() {
									goto l464
								}
								add(ruleError, position469)
							}
							break
						case '"':
							if !_rules[ruleString]() {
								goto l464
							}
							break
						case 'F', 'N', 'T', 'f', 'n', 't':
							{
								position470 := position
								{
									switch buffer[position] {
									case 'F', 'f':
										{
											position472, tokenIndex472 := position, tokenIndex
											if buffer[position] != rune('f') {
												goto l473
											}
											position++
											goto l472
										l473:
											position, tokenIndex = position472, tokenIndex472
											if buffer[position] != rune('F') {
												goto l464
											}
											position++
										}
									l472:
										{
											position474, tokenIndex474 := position, tokenIndex
											if buffer[position] != rune('a') {
												goto l475
											}
											position++
											goto l474
										l475:
											position, tokenIndex = position474, tokenIndex474
											if buffer[position] != rune('A') {
												goto l464
											}
											position++
										}
									l474:
										{
											position476, tokenIndex476 := position, tokenIndex
											if buffer[position] != rune('l') {
												goto l477
											}
											position++
											goto l476
										l477:
											position, tokenIndex = position476, tokenIndex476
											if buffer[position] != rune('L') {
												goto l464
											}
											position++
										}
									l476:
										{
											position478, tokenIndex478 := position, tokenIndex
											if buffer[position] != rune('s') {
												goto l479
											}
											position++
											goto l478
										l479:
											position, tokenIndex = position478, tokenIndex478
											if buffer[position] != rune('S') {
												goto l464
											}
											position++
										}
									l478:
										{
											position480, tokenIndex480 := position, tokenIndex
											if buffer[position] != rune('e') {
												goto l481
											}
											position++
											goto l480
										l481:
											position, tokenIndex = position480, tokenIndex480
											if buffer[position] != rune('E') {
												goto l464
											}
											position++
										}
									l480:
										break
									case 'T', 't':
										{
											position482, tokenIndex482 := position, tokenIndex
											if buffer[position] != rune('t') {
												goto l483
											}
											position++
											goto l482
										l483:
											position, tokenIndex = position482, tokenIndex482
											if buffer[position] != rune('T') {
												goto l464
											}
											position++
										}
									l482:
										{
											position484, tokenIndex484 := position, tokenIndex
											if buffer[position] != rune('r') {
												goto l485
											}
											position++
											goto l484
										l485:
											position, tokenIndex = position484, tokenIndex484
											if buffer[position] != rune('R') {
												goto l464
											}
											position++
										}
									l484:
										{
											position486, tokenIndex486 := position, tokenIndex
											if buffer[position] != rune('u') {
												goto l487
											}
											position++
											goto l486
										l487:
											position, tokenIndex = position486, tokenIndex486
											if buffer[position] != rune('U') {
												goto l464
											}
											position++
										}
									l486:
										{
											position488, tokenIndex488 := position, tokenIndex
											if buffer[position] != rune('e') {
												goto l489
											}
											position++
											goto l488
										l489:
											position, tokenIndex = position488, tokenIndex488
											if buffer[position] != rune('E') {
												goto l464
											}
											position++
										}
									l488:
										break
									default:
										{
											position490, tokenIndex490 := position, tokenIndex
											if buffer[position] != rune('n') {
												goto l491
											}
											position++
											goto l490
										l491:
											position, tokenIndex = position490, tokenIndex490
											if buffer[position] != rune('N') {
												goto l464
											}
											position++
										}
									l490:
										{
											position492, tokenIndex492 := position, tokenIndex
											if buffer[position] != rune('i') {
												goto l493
											}
											position++
											goto l492
										l493:
											position, tokenIndex = position492, tokenIndex492
											if buffer[position] != rune('I') {
												goto l464
											}
											position++
										}
									l492:
										{
											position494, tokenIndex494 := position, tokenIndex
											if buffer[position] != rune('l') {
												goto l495
											}
											position++
											goto l494
										l495:
											position, tokenIndex = position494, tokenIndex494
											if buffer[position] != rune('L') {
												goto l464
											}
											position++
										}
									l494:
										break
									}
								}

								add(ruleConstant, position470)
							}
							break
						case '[':
							{
								position496 := position
								if buffer[position] != rune('[') {
									goto l464
								}
								position++
							l497:
								{
									position498, tokenIndex498 := position, tokenIndex
									if !_rules[ruleExpression]() {
										goto l498
									}
									if buffer[position] != rune(',') {
										goto l498
									}
									position++
									{
										position499, tokenIndex499 := position, tokenIndex
										if !_rules[ruleWhitespace]() {
											goto l499
										}
										goto l500
									l499:
										position, tokenIndex = position499, tokenIndex499
									}
								l500:
									goto l497
								l498:
									position, tokenIndex = position498, tokenIndex498
								}
								{
									position501, tokenIndex501 := position, tokenIndex
									if !_rules[ruleExpression]() {
										goto l501
									}
									goto l502
								l501:
									position, tokenIndex = position501, tokenIndex501
								}
							l502:
								if buffer[position] != rune(']') {
									goto l464
								}
								position++
								add(ruleList, position496)
							}
							break
						default:
							{
								position503 := position
								{
									position504, tokenIndex504 := position, tokenIndex
									{
										position506 := position
										if c := buffer[position]; c < rune('0') || c > rune('9') {
											goto l505
										}
										position++
									l507:
										{
											position508, tokenIndex508 := position, tokenIndex
											if c := buffer[position]; c < rune('0') || c > rune('9') {
												goto l508
											}
											position++
											goto l507
										l508:
											position, tokenIndex = position508, tokenIndex508
										}
										if buffer[position] != rune('.') {
											goto l505
										}
										position++
										if c := buffer[position]; c < rune('0') || c > rune('9') {
											goto l505
										}
										position++
									l509:
										{
											position510, tokenIndex510 := position, tokenIndex
											if c := buffer[position]; c < rune('0') || c > rune('9') {
												goto l510
											}
											position++
											goto l509
										l510:
											position, tokenIndex = position510, tokenIndex510
										}
										add(ruleFloat, position506)
									}
									goto l504
								l505:
									position, tokenIndex = position504, tokenIndex504
									if !_rules[ruleInteger]() {
										goto l464
									}
								}
							l504:
								add(ruleNumber, position503)
							}
							break
						}
					}

				}
			l466:
				add(ruleSimple, position465)
			}
			return true
		l464:
			position, tokenIndex = position464, tokenIndex464
			return false
		},
		/* 38 List <- <('[' (Expression ',' Whitespace?)* Expression? ']')> */
		nil,
		/* 39 BinaryOperation <- <(ExpressionExceptBinaryOperation Whitespace BinaryOperator Whitespace ExpressionExceptBinaryOperation)> */
		func() bool {
			position512, tokenIndex512 := position, tokenIndex
			{
				position513 := position
				if !_rules[ruleExpressionExceptBinaryOperation]() {
					goto l512
				}
				if !_rules[ruleWhitespace]() {
					goto l512
				}
				{
					position514 := position
					{
						switch buffer[position] {
						case '/':
							if buffer[position] != rune('/') {
								goto l512
							}
							position++
							break
						case '*':
							if buffer[position] != rune('*') {
								goto l512
							}
							position++
							break
						case '-':
							if buffer[position] != rune('-') {
								goto l512
							}
							position++
							break
						default:
							if buffer[position] != rune('+') {
								goto l512
							}
							position++
							break
						}
					}

					add(ruleBinaryOperator, position514)
				}
				if !_rules[ruleWhitespace]() {
					goto l512
				}
				if !_rules[ruleExpressionExceptBinaryOperation]() {
					goto l512
				}
				add(ruleBinaryOperation, position513)
			}
			return true
		l512:
			position, tokenIndex = position512, tokenIndex512
			return false
		},
		/* 40 ExpressionExceptBinaryOperation <- <(Simple / UnaryOperation / Call)> */
		func() bool {
			position516, tokenIndex516 := position, tokenIndex
			{
				position517 := position
				{
					position518, tokenIndex518 := position, tokenIndex
					if !_rules[ruleSimple]() {
						goto l519
					}
					goto l518
				l519:
					position, tokenIndex = position518, tokenIndex518
					if !_rules[ruleUnaryOperation]() {
						goto l520
					}
					goto l518
				l520:
					position, tokenIndex = position518, tokenIndex518
					if !_rules[ruleCall]() {
						goto l516
					}
				}
			l518:
				add(ruleExpressionExceptBinaryOperation, position517)
			}
			return true
		l516:
			position, tokenIndex = position516, tokenIndex516
			return false
		},
		/* 41 BinaryOperator <- <((&('/') '/') | (&('*') '*') | (&('-') '-') | (&('+') '+'))> */
		nil,
		/* 42 UnaryOperation <- <(UnaryOperator ExpressionExceptOperation)> */
		func() bool {
			position522, tokenIndex522 := position, tokenIndex
			{
				position523 := position
				{
					position524 := position
					{
						position525, tokenIndex525 := position, tokenIndex
						if buffer[position] != rune('+') {
							goto l526
						}
						position++
						goto l525
					l526:
						position, tokenIndex = position525, tokenIndex525
						if buffer[position] != rune('-') {
							goto l522
						}
						position++
					}
				l525:
					add(ruleUnaryOperator, position524)
				}
				{
					position527 := position
					{
						position528, tokenIndex528 := position, tokenIndex
						if !_rules[ruleSimple]() {
							goto l529
						}
						goto l528
					l529:
						position, tokenIndex = position528, tokenIndex528
						if !_rules[ruleCall]() {
							goto l522
						}
					}
				l528:
					add(ruleExpressionExceptOperation, position527)
				}
				add(ruleUnaryOperation, position523)
			}
			return true
		l522:
			position, tokenIndex = position522, tokenIndex522
			return false
		},
		/* 43 UnaryOperator <- <('+' / '-')> */
		nil,
		/* 44 ExpressionExceptOperation <- <(Simple / Call)> */
		nil,
		/* 45 MethodCall <- <(Simple '.' Label '(' (Expression ',' Whitespace?)* Expression? ')')> */
		nil,
		/* 46 Call <- <(BuiltinCall / FunCall / MethodCall)> */
		func() bool {
			position533, tokenIndex533 := position, tokenIndex
			{
				position534 := position
				{
					position535, tokenIndex535 := position, tokenIndex
					{
						position537 := position
						{
							position538 := position
							{
								position539, tokenIndex539 := position, tokenIndex
								if buffer[position] != rune('m') {
									goto l540
								}
								position++
								goto l539
							l540:
								position, tokenIndex = position539, tokenIndex539
								if buffer[position] != rune('M') {
									goto l536
								}
								position++
							}
						l539:
							{
								position541, tokenIndex541 := position, tokenIndex
								if buffer[position] != rune('a') {
									goto l542
								}
								position++
								goto l541
							l542:
								position, tokenIndex = position541, tokenIndex541
								if buffer[position] != rune('A') {
									goto l536
								}
								position++
							}
						l541:
							{
								position543, tokenIndex543 := position, tokenIndex
								if buffer[position] != rune('k') {
									goto l544
								}
								position++
								goto l543
							l544:
								position, tokenIndex = position543, tokenIndex543
								if buffer[position] != rune('K') {
									goto l536
								}
								position++
							}
						l543:
							{
								position545, tokenIndex545 := position, tokenIndex
								if buffer[position] != rune('e') {
									goto l546
								}
								position++
								goto l545
							l546:
								position, tokenIndex = position545, tokenIndex545
								if buffer[position] != rune('E') {
									goto l536
								}
								position++
							}
						l545:
							add(ruleBuiltinFun, position538)
						}
						if buffer[position] != rune('(') {
							goto l536
						}
						position++
					l547:
						{
							position548, tokenIndex548 := position, tokenIndex
							if !_rules[ruleBuiltinArg]() {
								goto l548
							}
							if buffer[position] != rune(',') {
								goto l548
							}
							position++
							{
								position549, tokenIndex549 := position, tokenIndex
								if !_rules[ruleWhitespace]() {
									goto l549
								}
								goto l550
							l549:
								position, tokenIndex = position549, tokenIndex549
							}
						l550:
							goto l547
						l548:
							position, tokenIndex = position548, tokenIndex548
						}
						{
							position551, tokenIndex551 := position, tokenIndex
							if !_rules[ruleBuiltinArg]() {
								goto l551
							}
							goto l552
						l551:
							position, tokenIndex = position551, tokenIndex551
						}
					l552:
						if buffer[position] != rune(')') {
							goto l536
						}
						position++
						add(ruleBuiltinCall, position537)
					}
					goto l535
				l536:
					position, tokenIndex = position535, tokenIndex535
					{
						position554 := position
						if !_rules[ruleFunLabel]() {
							goto l553
						}
						if buffer[position] != rune('(') {
							goto l553
						}
						position++
					l555:
						{
							position556, tokenIndex556 := position, tokenIndex
							if !_rules[ruleExpression]() {
								goto l556
							}
							if buffer[position] != rune(',') {
								goto l556
							}
							position++
							{
								position557, tokenIndex557 := position, tokenIndex
								if !_rules[ruleWhitespace]() {
									goto l557
								}
								goto l558
							l557:
								position, tokenIndex = position557, tokenIndex557
							}
						l558:
							goto l555
						l556:
							position, tokenIndex = position556, tokenIndex556
						}
						{
							position559, tokenIndex559 := position, tokenIndex
							if !_rules[ruleExpression]() {
								goto l559
							}
							goto l560
						l559:
							position, tokenIndex = position559, tokenIndex559
						}
					l560:
						if buffer[position] != rune(')') {
							goto l553
						}
						position++
						add(ruleFunCall, position554)
					}
					goto l535
				l553:
					position, tokenIndex = position535, tokenIndex535
					{
						position561 := position
						if !_rules[ruleSimple]() {
							goto l533
						}
						if buffer[position] != rune('.') {
							goto l533
						}
						position++
						if !_rules[ruleLabel]() {
							goto l533
						}
						if buffer[position] != rune('(') {
							goto l533
						}
						position++
					l562:
						{
							position563, tokenIndex563 := position, tokenIndex
							if !_rules[ruleExpression]() {
								goto l563
							}
							if buffer[position] != rune(',') {
								goto l563
							}
							position++
							{
								position564, tokenIndex564 := position, tokenIndex
								if !_rules[ruleWhitespace]() {
									goto l564
								}
								goto l565
							l564:
								position, tokenIndex = position564, tokenIndex564
							}
						l565:
							goto l562
						l563:
							position, tokenIndex = position563, tokenIndex563
						}
						{
							position566, tokenIndex566 := position, tokenIndex
							if !_rules[ruleExpression]() {
								goto l566
							}
							goto l567
						l566:
							position, tokenIndex = position566, tokenIndex566
						}
					l567:
						if buffer[position] != rune(')') {
							goto l533
						}
						position++
						add(ruleMethodCall, position561)
					}
				}
			l535:
				add(ruleCall, position534)
			}
			return true
		l533:
			position, tokenIndex = position533, tokenIndex533
			return false
		},
		/* 47 BuiltinCall <- <(BuiltinFun '(' (BuiltinArg ',' Whitespace?)* BuiltinArg? ')')> */
		nil,
		/* 48 BuiltinFun <- <(('m' / 'M') ('a' / 'A') ('k' / 'K') ('e' / 'E'))> */
		nil,
		/* 49 BuiltinArg <- <(Type / Expression)> */
		func() bool {
			position570, tokenIndex570 := position, tokenIndex
			{
				position571 := position
				{
					position572, tokenIndex572 := position, tokenIndex
					if !_rules[ruleType]() {
						goto l573
					}
					goto l572
				l573:
					position, tokenIndex = position572, tokenIndex572
					if !_rules[ruleExpression]() {
						goto l570
					}
				}
			l572:
				add(ruleBuiltinArg, position571)
			}
			return true
		l570:
			position, tokenIndex = position570, tokenIndex570
			return false
		},
		/* 50 FunCall <- <(FunLabel '(' (Expression ',' Whitespace?)* Expression? ')')> */
		nil,
		/* 51 For <- <(ForIn / ForLoop)> */
		nil,
		/* 52 ForIn <- <(('f' / 'F') ('o' / 'O') ('r' / 'R') Whitespace (LowerLabel ',' Whitespace?)* LowerLabel Whitespace ('i' 'n') Whitespace Expression ':' Newline Indent Code)> */
		nil,
		/* 53 ForLoop <- <(('f' / 'F') ('o' / 'O') ('r' / 'R') Whitespace LowerLabel Whitespace ('i' 'n') Whitespace Range ':' Newline Indent Code)> */
		nil,
		/* 54 Range <- <(Integer RangeOperator Integer)> */
		nil,
		/* 55 RangeOperator <- <(('.' '.' '.') / ('.' '.'))> */
		nil,
		/* 56 On <- <(('o' / 'O') ('n' / 'N') Whitespace FunLabel ':' Newline Indent Code)> */
		nil,
		/* 57 Return <- <((&('E' | 'e') Escalator) | (&('!') ReturnError) | (&('R' | 'r') ReturnValue))> */
		nil,
		/* 58 ReturnValue <- <(('r' / 'R') ('e' / 'E') ('t' / 'T') ('u' / 'U') ('r' / 'R') ('n' / 'N') Whitespace? Expression)> */
		nil,
		/* 59 ReturnError <- <('!' '!' Whitespace? Expression)> */
		nil,
		/* 60 Escalator <- <(('e' / 'E') ('s' / 'S') ('c' / 'C') ('a' / 'A') ('l' / 'L') ('a' / 'A') ('t' / 'T') ('e' / 'E') Whitespace (FunLabel ',' Whitespace?)* FunLabel)> */
		nil,
		/* 61 LowerLabel <- <([a-z] ((&('_') '_') | (&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') [0-9]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))*)> */
		func() bool {
			position585, tokenIndex585 := position, tokenIndex
			{
				position586 := position
				if c := buffer[position]; c < rune('a') || c > rune('z') {
					goto l585
				}
				position++
			l587:
				{
					position588, tokenIndex588 := position, tokenIndex
					{
						switch buffer[position] {
						case '_':
							if buffer[position] != rune('_') {
								goto l588
							}
							position++
							break
						case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l588
							}
							position++
							break
						default:
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l588
							}
							position++
							break
						}
					}

					goto l587
				l588:
					position, tokenIndex = position588, tokenIndex588
				}
				add(ruleLowerLabel, position586)
			}
			return true
		l585:
			position, tokenIndex = position585, tokenIndex585
			return false
		},
		/* 62 CapitalLabel <- <([A-Z] ((&('_') '_') | (&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') [0-9]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]))*)> */
		func() bool {
			position590, tokenIndex590 := position, tokenIndex
			{
				position591 := position
				if c := buffer[position]; c < rune('A') || c > rune('Z') {
					goto l590
				}
				position++
			l592:
				{
					position593, tokenIndex593 := position, tokenIndex
					{
						switch buffer[position] {
						case '_':
							if buffer[position] != rune('_') {
								goto l593
							}
							position++
							break
						case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l593
							}
							position++
							break
						case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z':
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l593
							}
							position++
							break
						default:
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l593
							}
							position++
							break
						}
					}

					goto l592
				l593:
					position, tokenIndex = position593, tokenIndex593
				}
				add(ruleCapitalLabel, position591)
			}
			return true
		l590:
			position, tokenIndex = position590, tokenIndex590
			return false
		},
		/* 63 FunLowerLabel <- <([a-z] ((&('_') '_') | (&('`') '`') | (&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') [0-9]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))* ('?' / '!')?)> */
		func() bool {
			position595, tokenIndex595 := position, tokenIndex
			{
				position596 := position
				if c := buffer[position]; c < rune('a') || c > rune('z') {
					goto l595
				}
				position++
			l597:
				{
					position598, tokenIndex598 := position, tokenIndex
					{
						switch buffer[position] {
						case '_':
							if buffer[position] != rune('_') {
								goto l598
							}
							position++
							break
						case '`':
							if buffer[position] != rune('`') {
								goto l598
							}
							position++
							break
						case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l598
							}
							position++
							break
						default:
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l598
							}
							position++
							break
						}
					}

					goto l597
				l598:
					position, tokenIndex = position598, tokenIndex598
				}
				{
					position600, tokenIndex600 := position, tokenIndex
					{
						position602, tokenIndex602 := position, tokenIndex
						if buffer[position] != rune('?') {
							goto l603
						}
						position++
						goto l602
					l603:
						position, tokenIndex = position602, tokenIndex602
						if buffer[position] != rune('!') {
							goto l600
						}
						position++
					}
				l602:
					goto l601
				l600:
					position, tokenIndex = position600, tokenIndex600
				}
			l601:
				add(ruleFunLowerLabel, position596)
			}
			return true
		l595:
			position, tokenIndex = position595, tokenIndex595
			return false
		},
		/* 64 Label <- <(FunLabel / CapitalLabel / LowerLabel)> */
		func() bool {
			position604, tokenIndex604 := position, tokenIndex
			{
				position605 := position
				{
					position606, tokenIndex606 := position, tokenIndex
					if !_rules[ruleFunLabel]() {
						goto l607
					}
					goto l606
				l607:
					position, tokenIndex = position606, tokenIndex606
					if !_rules[ruleCapitalLabel]() {
						goto l608
					}
					goto l606
				l608:
					position, tokenIndex = position606, tokenIndex606
					if !_rules[ruleLowerLabel]() {
						goto l604
					}
				}
			l606:
				add(ruleLabel, position605)
			}
			return true
		l604:
			position, tokenIndex = position604, tokenIndex604
			return false
		},
		/* 65 Float <- <([0-9]+ '.' [0-9]+)> */
		nil,
		/* 66 Integer <- <[0-9]+> */
		func() bool {
			position610, tokenIndex610 := position, tokenIndex
			{
				position611 := position
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l610
				}
				position++
			l612:
				{
					position613, tokenIndex613 := position, tokenIndex
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l613
					}
					position++
					goto l612
				l613:
					position, tokenIndex = position613, tokenIndex613
				}
				add(ruleInteger, position611)
			}
			return true
		l610:
			position, tokenIndex = position610, tokenIndex610
			return false
		},
		/* 67 Number <- <(Float / Integer)> */
		nil,
		/* 68 Constant <- <((&('F' | 'f') (('f' / 'F') ('a' / 'A') ('l' / 'L') ('s' / 'S') ('e' / 'E'))) | (&('T' | 't') (('t' / 'T') ('r' / 'R') ('u' / 'U') ('e' / 'E'))) | (&('N' | 'n') (('n' / 'N') ('i' / 'I') ('l' / 'L'))))> */
		nil,
		/* 69 String <- <(Template / Text)> */
		func() bool {
			position616, tokenIndex616 := position, tokenIndex
			{
				position617 := position
				{
					position618, tokenIndex618 := position, tokenIndex
					{
						position620 := position
						if buffer[position] != rune('"') {
							goto l619
						}
						position++
						{
							position623 := position
						l624:
							{
								position625, tokenIndex625 := position, tokenIndex
								{
									position626, tokenIndex626 := position, tokenIndex
									if buffer[position] != rune('#') {
										goto l626
									}
									position++
									goto l625
								l626:
									position, tokenIndex = position626, tokenIndex626
								}
								if !matchDot() {
									goto l625
								}
								goto l624
							l625:
								position, tokenIndex = position625, tokenIndex625
							}
							add(ruleSegment, position623)
						}
						{
							position627 := position
							if buffer[position] != rune('#') {
								goto l619
							}
							position++
							if buffer[position] != rune('{') {
								goto l619
							}
							position++
							if !_rules[ruleExpression]() {
								goto l619
							}
							if buffer[position] != rune('}') {
								goto l619
							}
							position++
							add(ruleSlot, position627)
						}
					l621:
						{
							position622, tokenIndex622 := position, tokenIndex
							{
								position628 := position
							l629:
								{
									position630, tokenIndex630 := position, tokenIndex
									{
										position631, tokenIndex631 := position, tokenIndex
										if buffer[position] != rune('#') {
											goto l631
										}
										position++
										goto l630
									l631:
										position, tokenIndex = position631, tokenIndex631
									}
									if !matchDot() {
										goto l630
									}
									goto l629
								l630:
									position, tokenIndex = position630, tokenIndex630
								}
								add(ruleSegment, position628)
							}
							{
								position632 := position
								if buffer[position] != rune('#') {
									goto l622
								}
								position++
								if buffer[position] != rune('{') {
									goto l622
								}
								position++
								if !_rules[ruleExpression]() {
									goto l622
								}
								if buffer[position] != rune('}') {
									goto l622
								}
								position++
								add(ruleSlot, position632)
							}
							goto l621
						l622:
							position, tokenIndex = position622, tokenIndex622
						}
						{
							position633 := position
						l634:
							{
								position635, tokenIndex635 := position, tokenIndex
								{
									position636, tokenIndex636 := position, tokenIndex
									if buffer[position] != rune('"') {
										goto l636
									}
									position++
									goto l635
								l636:
									position, tokenIndex = position636, tokenIndex636
								}
								if !matchDot() {
									goto l635
								}
								goto l634
							l635:
								position, tokenIndex = position635, tokenIndex635
							}
							add(ruleQ, position633)
						}
						if buffer[position] != rune('"') {
							goto l619
						}
						position++
						add(ruleTemplate, position620)
					}
					goto l618
				l619:
					position, tokenIndex = position618, tokenIndex618
					{
						position637 := position
						if buffer[position] != rune('"') {
							goto l616
						}
						position++
					l638:
						{
							position639, tokenIndex639 := position, tokenIndex
							{
								position640, tokenIndex640 := position, tokenIndex
								if buffer[position] != rune('"') {
									goto l640
								}
								position++
								goto l639
							l640:
								position, tokenIndex = position640, tokenIndex640
							}
							if !matchDot() {
								goto l639
							}
							goto l638
						l639:
							position, tokenIndex = position639, tokenIndex639
						}
						if buffer[position] != rune('"') {
							goto l616
						}
						position++
						add(ruleText, position637)
					}
				}
			l618:
				add(ruleString, position617)
			}
			return true
		l616:
			position, tokenIndex = position616, tokenIndex616
			return false
		},
		/* 70 Template <- <('"' (Segment Slot)+ Q '"')> */
		nil,
		/* 71 Segment <- <(!'#' .)*> */
		nil,
		/* 72 Q <- <(!'"' .)*> */
		nil,
		/* 73 Text <- <('"' (!'"' .)* '"')> */
		nil,
		/* 74 Error <- <('$' Label)> */
		nil,
		/* 75 Slot <- <('#' '{' Expression '}')> */
		nil,
		/* 76 Whitespace <- <' '+> */
		func() bool {
			position647, tokenIndex647 := position, tokenIndex
			{
				position648 := position
				if buffer[position] != rune(' ') {
					goto l647
				}
				position++
			l649:
				{
					position650, tokenIndex650 := position, tokenIndex
					if buffer[position] != rune(' ') {
						goto l650
					}
					position++
					goto l649
				l650:
					position, tokenIndex = position650, tokenIndex650
				}
				add(ruleWhitespace, position648)
			}
			return true
		l647:
			position, tokenIndex = position647, tokenIndex647
			return false
		},
		/* 77 Newline <- <'\n'+> */
		func() bool {
			position651, tokenIndex651 := position, tokenIndex
			{
				position652 := position
				if buffer[position] != rune('\n') {
					goto l651
				}
				position++
			l653:
				{
					position654, tokenIndex654 := position, tokenIndex
					if buffer[position] != rune('\n') {
						goto l654
					}
					position++
					goto l653
				l654:
					position, tokenIndex = position654, tokenIndex654
				}
				add(ruleNewline, position652)
			}
			return true
		l651:
			position, tokenIndex = position651, tokenIndex651
			return false
		},
		/* 78 EOT <- <!.> */
		nil,
	}
	p.rules = _rules
}
