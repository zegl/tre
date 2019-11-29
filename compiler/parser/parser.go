package parser

import (
	"fmt"
	"log"
	"strconv"

	"errors"

	"github.com/zegl/tre/compiler/lexer"
)

type parser struct {
	i     int
	input []lexer.Item

	inAllocRightHand bool

	debug bool

	types map[string]struct{}
}

func Parse(input []lexer.Item, debug bool) FileNode {
	p := &parser{
		i:     0,
		input: input,
		debug: debug,
		types: map[string]struct{}{
			"int":   {},
			"int8":  {},
			"int32": {},
			"int64": {},
		},
	}

	return FileNode{
		Instructions: p.parseUntil(lexer.Item{Type: lexer.EOF}),
	}
}

func (p *parser) parseOne(withAheadParse bool) (res Node) {
	return p.parseOneWithOptions(withAheadParse, withAheadParse, withAheadParse)
}

func (p *parser) parseOneWithOptions(withAheadParse, withArithAhead, withIdentifierAhead bool) (res Node) {
	current := p.input[p.i]

	switch current.Type {

	case lexer.EOF:
		panic("unexpected EOF")

	case lexer.EOL:
		// Ignore the EOL, continue further
		// p.i++
		// return p.parseOne()
		return nil

	// IDENTIFIERS are converted to either:
	// - a CallNode if followed by an opening parenthesis (a function call), or
	// - a NodeName (variables)
	case lexer.IDENTIFIER:
		res = &NameNode{Name: current.Val}
		if withIdentifierAhead {
			res = p.aheadParseWithOptions(res, withArithAhead, withIdentifierAhead)
		}
		return

		// NUMBER always returns a ConstantNode
		// Convert string representation to int64
	case lexer.NUMBER:
		val, err := strconv.ParseInt(current.Val, 10, 64)
		if err != nil {
			panic(err)
		}

		res = &ConstantNode{
			Type:  NUMBER,
			Value: val,
		}
		if withAheadParse {
			res = p.aheadParse(res)
		}
		return

		// STRING is always a ConstantNode, the value is not modified
	case lexer.STRING:
		res = &ConstantNode{
			Type:     STRING,
			ValueStr: current.Val,
		}
		if withAheadParse {
			res = p.aheadParse(res)
		}
		return

	case lexer.OPERATOR:
		if current.Val == "&" {
			p.i++
			res = &GetReferenceNode{Item: p.parseOne(true)}
			if withAheadParse {
				res = p.aheadParse(res)
			}
			return
		}
		if current.Val == "*" {
			p.i++
			res = &DereferenceNode{Item: p.parseOne(false)}
			if withAheadParse {
				res = p.aheadParse(res)
			}
			return
		}

		if current.Val == "!" {
			p.i++
			res = &NegateNode{Item: p.parseOne(false)}
			return
		}

		if current.Val == "-" {
			p.i++
			res = p.aheadParse(&SubNode{Item: p.parseOne(true)})
			return
		}

		// Slice or array initalization
		if current.Val == "[" {
			next := p.lookAhead(1)

			// Slice init
			if next.Type == lexer.OPERATOR && next.Val == "]" {
				p.i += 2

				sliceItemType, err := p.parseOneType()
				if err != nil {
					panic(err)
				}

				p.i++

				next = p.lookAhead(0)
				if next.Type != lexer.SEPARATOR || next.Val != "{" {
					log.Printf("%+v", next)
					panic("expected { after type in slice init")
				}

				p.i++

				prevInAlloc := p.inAllocRightHand
				p.inAllocRightHand = false
				items := p.parseUntil(lexer.Item{Type: lexer.SEPARATOR, Val: "}"})
				p.inAllocRightHand = prevInAlloc

				res = &InitializeSliceNode{
					Type:  sliceItemType,
					Items: items,
				}
				if withAheadParse {
					res = p.aheadParse(res)
				}
				return
			}

			p.i++

			// TODO: Support for compile-time artimethic ("[1+2]int{1,2,3}")
			arraySize := p.lookAhead(0)
			if arraySize.Type != lexer.NUMBER {
				panic("expected number in array size")
			}
			size, err := strconv.Atoi(arraySize.Val)
			if err != nil {
				panic("expected number in array size")
			}

			p.i++
			p.expect(p.lookAhead(0), lexer.Item{
				Type: lexer.OPERATOR,
				Val:  "]",
			})

			p.i++
			arrayItemType, err := p.parseOneType()
			if err != nil {
				panic(err)
			}

			p.i++
			p.expect(p.lookAhead(0), lexer.Item{
				Type: lexer.SEPARATOR,
				Val:  "{",
			})

			p.i++

			prevInAlloc := p.inAllocRightHand
			p.inAllocRightHand = false
			items := p.parseUntil(lexer.Item{Type: lexer.SEPARATOR, Val: "}"})
			p.inAllocRightHand = prevInAlloc

			// Array init
			res = &InitializeArrayNode{
				Type:  arrayItemType,
				Size:  size,
				Items: items,
			}
			if withAheadParse {
				res = p.aheadParse(res)
			}
			return
		}

	case lexer.KEYWORD:

		// "if" gets converted to a ConditionNode
		// the keyword "if" is followed by
		// - a condition
		// - an opening curly bracket ({)
		// - a body
		// - a closing bracket (})
		if current.Val == "if" {

			getCondition := func() *OperatorNode {
				p.i++

				condNodes := p.parseUntil(lexer.Item{Type: lexer.SEPARATOR, Val: "{"})
				if len(condNodes) != 1 {
					panic("could not parse if-condition")
				}

				p.i++

				if cond, ok := condNodes[0].(*OperatorNode); ok {
					return cond
				}

				// Add implicit == true
				return &OperatorNode{
					Left: condNodes[0],
					Right: &ConstantNode{
						Type:  BOOL,
						Value: 1,
					},
					Operator: OP_EQ,
				}

			}

			outerConditionNode := &ConditionNode{
				Cond: getCondition(),
				True: p.parseUntil(lexer.Item{Type: lexer.SEPARATOR, Val: "}"}),
			}

			lastConditionNode := outerConditionNode

			p.i++

			// Check if the next keyword is "if" + "else" or "else"
			for {
				checkIfElse := p.lookAhead(0)
				if checkIfElse.Type != lexer.KEYWORD || checkIfElse.Val != "else" {
					break
				}

				p.i++

				checkIfElseIf := p.lookAhead(0)
				if checkIfElseIf.Type == lexer.KEYWORD && checkIfElseIf.Val == "if" {

					newCondNode := &ConditionNode{
						Cond: getCondition(),
						True: p.parseUntil(lexer.Item{Type: lexer.SEPARATOR, Val: "}"}),
					}

					lastConditionNode.False = []Node{newCondNode}
					lastConditionNode = newCondNode
					p.i++
					continue
				}

				expectOpenBrack := p.lookAhead(0)
				if expectOpenBrack.Type != lexer.SEPARATOR || expectOpenBrack.Val != "{" {
					panic("Expected { after else")
				}

				p.i++
				lastConditionNode.False = p.parseUntil(lexer.Item{Type: lexer.SEPARATOR, Val: "}"})
			}

			return outerConditionNode
		}

		// "func" gets converted into a DefineFuncNode
		// the keyword "func" is followed by
		// - a IDENTIFIER (function name)
		// - opening parenthesis
		// - optional: arguments (name type, name2 type2, ...)
		// - closing parenthesis
		// - optional: return type
		// - opening curly bracket ({)
		// - function body
		// - closing curly bracket (})

		// named func:  func abc() {
		// method:      func (a abc) abc() {
		// value func:  func (a abc) {

		if current.Val == "func" {
			defineFunc := &DefineFuncNode{}
			p.i++

			var argsOrMethodType []*NameNode
			var canBeMethod bool

			checkIfOpeningParen := p.lookAhead(0)
			if checkIfOpeningParen.Type == lexer.SEPARATOR && checkIfOpeningParen.Val == "(" {
				p.i++
				argsOrMethodType = p.parseFunctionArguments()
				canBeMethod = true
			}

			checkIfIdentifier := p.lookAhead(0)
			checkIfOpeningParen = p.lookAhead(1)

			if canBeMethod && checkIfIdentifier.Type == lexer.IDENTIFIER &&
				checkIfOpeningParen.Type == lexer.SEPARATOR && checkIfOpeningParen.Val == "(" {

				defineFunc.IsMethod = true
				defineFunc.IsNamed = true
				defineFunc.Name = checkIfIdentifier.Val

				if len(argsOrMethodType) != 1 {
					panic("Unexpected count of types in method")
				}

				defineFunc.InstanceName = argsOrMethodType[0].Name

				methodOnType := argsOrMethodType[0].Type

				if pointerSingleTypeNode, ok := methodOnType.(*PointerTypeNode); ok {
					defineFunc.IsPointerReceiver = true
					methodOnType = pointerSingleTypeNode.ValueType
				}

				if singleTypeNode, ok := methodOnType.(*SingleTypeNode); ok {
					defineFunc.MethodOnType = singleTypeNode
				} else {
					panic(fmt.Sprintf("could not find type in method defitition: %T", methodOnType))
				}
			}

			name := p.lookAhead(0)
			openParen := p.lookAhead(1)
			if name.Type == lexer.IDENTIFIER && openParen.Type == lexer.SEPARATOR && openParen.Val == "(" {
				defineFunc.Name = name.Val
				defineFunc.IsNamed = true

				p.i++
				p.i++

				// Parse argument list
				defineFunc.Arguments = p.parseFunctionArguments()
			} else {
				defineFunc.Arguments = argsOrMethodType
			}

			// Parse return types
			var retTypesNodeNames []*NameNode

			checkIfOpeningCurly := p.lookAhead(0)
			if checkIfOpeningCurly.Type != lexer.SEPARATOR || checkIfOpeningCurly.Val != "{" {

				// Check if next is an opening parenthesis
				// Is optional if there's only one return type, is required
				// if there is multiple ones
				var allowMultiRetVals bool

				checkIfOpenParen := p.lookAhead(0)
				if checkIfOpenParen.Type == lexer.SEPARATOR && checkIfOpenParen.Val == "(" {
					allowMultiRetVals = true
					p.i++
				}

				for {
					nameNode := &NameNode{}

					// Support both named return values and when we only get the type
					retTypeOrNamed, err := p.parseOneType()
					if err != nil {
						panic(err)
					}
					p.i++

					// Next can be type, that means that the previous was the name of the var
					isType := p.lookAhead(0)
					if isType.Type == lexer.IDENTIFIER {
						retType, err := p.parseOneType()
						if err != nil {
							panic(err)
						}
						p.i++

						nameNode.Name = retTypeOrNamed.Type()
						nameNode.Type = retType
					} else {
						nameNode.Type = retTypeOrNamed
					}

					retTypesNodeNames = append(retTypesNodeNames, nameNode)

					if !allowMultiRetVals {
						break
					}

					// Check if comma or end parenthesis
					commaOrEnd := p.lookAhead(0)
					if commaOrEnd.Type == lexer.SEPARATOR && commaOrEnd.Val == "," {
						p.i++
						continue
					}

					if commaOrEnd.Type == lexer.SEPARATOR && commaOrEnd.Val == ")" {
						p.i++
						break
					}
					panic("Could not parse function return types")
				}
			}

			defineFunc.ReturnValues = retTypesNodeNames

			openBracket := p.lookAhead(0)
			if openBracket.Type != lexer.SEPARATOR || openBracket.Val != "{" {
				panic("func arguments must be followed by {. Got " + openBracket.Val)
			}

			p.i++

			defineFunc.Body = p.parseUntil(lexer.Item{Type: lexer.SEPARATOR, Val: "}"})

			return p.aheadParse(defineFunc)
		}

		// "return" creates a ReturnNode
		if current.Val == "return" {
			p.i++

			var retVals []Node

			for {
				checkIfEOL := p.lookAhead(0)
				if checkIfEOL.Type == lexer.EOL {
					break
				}

				retVals = append(retVals, p.parseOne(true))
				p.i++

				checkIfComma := p.lookAhead(0)
				if checkIfComma.Type == lexer.SEPARATOR && checkIfComma.Val == "," {
					p.i++
					continue
				}

				break
			}

			res = &ReturnNode{Vals: retVals}
			if withAheadParse {
				res = p.aheadParse(res)
			}
			return
		}

		// Declare a new type
		if current.Val == "type" {
			name := p.lookAhead(1)
			if name.Type != lexer.IDENTIFIER {
				panic("type must beb followed by IDENTIFIER")
			}

			p.i += 2

			typeType, err := p.parseOneType()
			if err != nil {
				panic(err)
			}

			// Save the name of the type
			typeType.SetName(name.Val)

			res = &DefineTypeNode{
				Name: name.Val,
				Type: typeType,
			}
			if withAheadParse {
				res = p.aheadParse(res)
			}

			// Register that this type exists
			// TODO: Make it context sensitive (eg package level types, types in functions etc)
			p.types[name.Val] = struct{}{}

			return
		}

		// New instance of type
		if current.Val == "var" {
			p.i++

			name := p.lookAhead(0)
			if name.Type != lexer.IDENTIFIER {
				panic("expected IDENTIFIER after var")
			}

			p.i++

			tp, err := p.parseOneType()
			if err != nil {
				panic(err)
			}

			return &AllocNode{
				Name: []string{name.Val},
				Val:  []Node{tp},
			}
		}

		if current.Val == "package" {
			packageName := p.lookAhead(1)

			if packageName.Type != lexer.IDENTIFIER {
				panic("package must be followed by a IDENTIFIER")
			}

			p.i += 1

			return &DeclarePackageNode{
				PackageName: packageName.Val,
			}
		}

		if current.Val == "for" {
			return p.parseFor()
		}

		if current.Val == "break" {
			return &BreakNode{}
		}

		if current.Val == "continue" {
			return &ContinueNode{}
		}

		if current.Val == "import" {
			return p.parseImport()
		}

		if current.Val == "true" || current.Val == "false" {
			var v int64 = 0
			if current.Val == "true" {
				v = 1
			}

			return &ConstantNode{
				Type:  BOOL,
				Value: v,
			}
		}

		if current.Val == "range" {
			return p.parseRange()
		}

		if current.Val == "switch" {
			return p.parseSwitch()
		}
	}

	p.printInput()
	log.Panicf("unable to handle default: %d - %+v", p.i, current)
	panic("")
}

func (p *parser) aheadParse(input Node) Node {
	return p.aheadParseWithOptions(input, true, true)
}

func (p *parser) aheadParseWithOptions(input Node, withArithAhead, withIdentifierAhead bool) Node {
	next := p.lookAhead(1)

	if next.Type == lexer.OPERATOR {
		if next.Val == "." {
			p.i++

			next = p.lookAhead(1)
			if next.Type == lexer.IDENTIFIER {
				p.i++
				return p.aheadParseWithOptions(&StructLoadElementNode{
					Struct:      input,
					ElementName: next.Val,
				}, withArithAhead, withIdentifierAhead)
			}

			if next.Type == lexer.SEPARATOR && next.Val == "(" {
				p.i++
				p.i++

				castToType, err := p.parseOneType()
				if err != nil {
					panic(err)
				}

				p.i++

				expectEndParen := p.lookAhead(0)
				p.expect(expectEndParen, lexer.Item{Type: lexer.SEPARATOR, Val: ")"})

				p.i++

				return p.aheadParse(&TypeCastInterfaceNode{
					Item: input,
					Type: castToType,
				})
			}

			panic(fmt.Sprintf("Expected IDENTFIER or ( after . Got: %+v", next))
		}

		if next.Val == ":=" || next.Val == "=" {
			p.i += 2

			// TODO: This needs to be a stack
			prevInRight := p.inAllocRightHand
			val := p.parseOne(true)
			p.inAllocRightHand = prevInRight

			if nameNode, ok := input.(*NameNode); ok {
				if next.Val == ":=" {
					return &AllocNode{
						Name: []string{nameNode.Name},
						Val:  []Node{val},
					}
				} else {
					return &AssignNode{
						Target: []Node{nameNode},
						Val:    []Node{val},
					}
				}
			}

			if next.Val == "=" {
				return &AssignNode{
					Target: []Node{input},
					Val:    []Node{val},
				}
			}

			panic(fmt.Sprintf("%s can only be used after a name. Got: %+v", next.Val, input))
		}

		if next.Val == "+=" || next.Val == "-=" || next.Val == "*=" || next.Val == "/=" {
			p.i++
			p.i++

			var op Operator
			switch next.Val {
			case "+=":
				op = OP_ADD
			case "-=":
				op = OP_SUB
			case "*=":
				op = OP_MUL
			case "/=":
				op = OP_DIV
			}

			return &AssignNode{
				Target: []Node{input},
				Val: []Node{&OperatorNode{
					Operator: op,
					Left:     input,
					Right:    p.parseOne(true),
				}},
			}
		}

		if next.Val == "..." {
			p.i++
			return &DeVariadicSliceNode{
				Item: input,
			}
		}

		// Array slicing
		if next.Val == "[" {
			p.i += 2

			index := p.parseOne(true)

			var res Node

			checkIfColon := p.lookAhead(1)
			if checkIfColon.Type == lexer.OPERATOR && checkIfColon.Val == ":" {
				p.i += 2
				res = &SliceArrayNode{
					Val:    input,
					Start:  index,
					HasEnd: true,
					End:    p.parseOne(true),
				}
			} else {
				res = &LoadArrayElement{
					Array: input,
					Pos:   index,
				}
			}

			p.i++

			expectEndBracket := p.lookAhead(0)
			if expectEndBracket.Type == lexer.OPERATOR && expectEndBracket.Val == "]" {
				return p.aheadParse(res)
			}

			panic(fmt.Sprintf("Unexpected %+v, expected ]", expectEndBracket))
		}

		// Handle "Operations" both arith and comparision
		if _, ok := opsCharToOp[next.Val]; ok {
			operator := opsCharToOp[next.Val]
			_, isArithOp := arithOperators[operator]

			if !withArithAhead && isArithOp {
				return input
			}

			p.i += 2
			res := &OperatorNode{
				Operator: operator,
				Left:     input,
			}

			if isArithOp {
				res.Right = p.parseOneWithOptions(false, false, true)
				// Sort infix operations if necessary (eg: apply OP_MUL before OP_ADD)
				res = sortInfix(res)
			} else {
				res.Right = p.parseOneWithOptions(true, true, true)
			}

			return p.aheadParseWithOptions(res, true, true)
		}

		if next.Val == "--" {
			p.i++
			return &DecrementNode{Item: input}
		}

		if next.Val == "++" {
			p.i++
			return &IncrementNode{Item: input}
		}
	}

	if next.Type == lexer.SEPARATOR && next.Val == "(" {
		current := p.lookAhead(0)

		p.i += 2 // identifier and left paren

		if _, ok := p.types[current.Val]; ok {
			val := p.parseUntil(lexer.Item{Type: lexer.SEPARATOR, Val: ")"})
			if len(val) != 1 {
				panic("type conversion must take only one argument")
			}
			return p.aheadParse(&TypeCastNode{
				Type: &SingleTypeNode{
					TypeName: current.Val,
				},
				Val: val[0],
			})
		}

		beforeAllocRightHand := p.inAllocRightHand
		p.inAllocRightHand = false
		callNode := p.aheadParse(&CallNode{
			Function:  input,
			Arguments: p.parseUntil(lexer.Item{Type: lexer.SEPARATOR, Val: ")"}),
		})
		p.inAllocRightHand = beforeAllocRightHand
		return callNode
	}

	// Initialize structs with values:
	//   Foo{Bar: 123}
	//   Foo{Bar: 123, Bax: hello(123)}
	if next.Type == lexer.SEPARATOR && next.Val == "{" {
		nameNode, isNamedNode := input.(*NameNode)
		if isNamedNode {
			_, isType := p.types[nameNode.Name]
			if isType {
				inputType := &SingleTypeNode{
					TypeName: nameNode.Name,
				}

				p.i += 2

				items := make(map[string]Node)

				for {
					// Skip EOLs
					checkIfEOL := p.lookAhead(0)
					if checkIfEOL.Type == lexer.EOL {
						p.i++
					}

					// Find end of parsing
					checkIfEndBracket := p.lookAhead(0)
					if checkIfEndBracket.Type == lexer.SEPARATOR && checkIfEndBracket.Val == "}" {
						p.i++
						break
					}

					key := p.lookAhead(0)
					if key.Type != lexer.IDENTIFIER {
						panic("Expected IDENTIFIER in struct initialization")
					}

					col := p.lookAhead(1)
					p.expect(col, lexer.Item{Type: lexer.OPERATOR, Val: ":"})

					p.i += 2

					prevInAlloc := p.inAllocRightHand
					p.inAllocRightHand = false
					items[key.Val] = p.parseOne(true)
					p.inAllocRightHand = prevInAlloc

					p.i++

					commaOrEnd := p.lookAhead(0)
					if commaOrEnd.Type == lexer.SEPARATOR && commaOrEnd.Val == "," {
						p.i++
						continue
					}

					if commaOrEnd.Type == lexer.SEPARATOR && commaOrEnd.Val == "}" {
						break
					}
				}

				return p.aheadParse(&InitializeStructNode{
					Type:  inputType,
					Items: items,
				})
			}
		}
	}

	if next.Type == lexer.SEPARATOR && next.Val == "," {
		// MultiName node parsing ("a, b, c := ...")
		if inputNamedNode, ok := input.(*NameNode); ok {

			// This bit of parsing is specualtive
			//
			// The parser will restore it's position to this index in case it
			// turns out that we can not convert this into a MultiNameNode
			preIndex := p.i

			p.i++
			p.i++

			nextName := p.parseOne(true)

			if nextAlloc, ok := nextName.(*AllocNode); ok {
				nextNames := []string{inputNamedNode.Name}
				nextNames = append(nextNames, nextAlloc.Name...)
				nextAlloc.Name = nextNames

				prev := p.inAllocRightHand
				p.inAllocRightHand = true
				r := p.aheadParse(nextAlloc)
				p.inAllocRightHand = prev

				return r
			}

			if nextAssign, ok := nextName.(*AssignNode); ok {
				nextTargets := []Node{input}
				nextTargets = append(nextTargets, nextAssign.Target...)
				nextAssign.Target = nextTargets

				prev := p.inAllocRightHand
				p.inAllocRightHand = true
				r := p.aheadParse(nextAssign)
				p.inAllocRightHand = prev

				return r
			}

			// A MultiNameNode could not be created
			// Reset the parsing index
			p.i = preIndex
		}

		// MultiValueNode node parsing ("... := 1, 2, 3")
		if p.inAllocRightHand {
			p.i += 2

			// Add to existing multi value if possible. Done on third argument
			// and forward
			if inAllocNode, ok := input.(*AllocNode); ok {
				inAllocNode.Val = append(inAllocNode.Val, p.parseOne(false))
				return p.aheadParse(inAllocNode)
			}

			if inAssignValue, ok := input.(*AssignNode); ok {
				inAssignValue.Val = append(inAssignValue.Val, p.parseOne(false))
				return p.aheadParse(inAssignValue)
			}

			panic("unexpected in alloc right hand")
		}
	}

	return input
}

func (p *parser) lookAhead(steps int) lexer.Item {
	return p.input[p.i+steps]
}

func (p *parser) parseUntil(until lexer.Item) []Node {
	n, _ := p.parseUntilEither([]lexer.Item{until})
	return n
}

// parseUntilEither reads lexer items until it finds one that equals to a item in "untils"
// The list of parsed nodes is returned in res. The lexer item that stopped the iteration
// is returned in "reached"
func (p *parser) parseUntilEither(untils []lexer.Item) (res []Node, reached lexer.Item) {
	for {
		current := p.input[p.i]

		// Check if we have reached the end
		for _, until := range untils {
			if current.Type == until.Type && current.Val == until.Val {
				return res, until
			}
		}

		// Ignore comma
		if current.Type == lexer.SEPARATOR && current.Val == "," {
			p.i++
			continue
		}

		one := p.parseOne(true)
		if one != nil {
			res = append(res, one)
		}

		p.i++
	}
}

func (p *parser) parseFunctionArguments() []*NameNode {
	var res []*NameNode
	var i int

	for {
		current := p.input[p.i]
		if current.Type == lexer.SEPARATOR && current.Val == ")" {
			p.i++
			return res
		}

		if i > 0 {
			if current.Type != lexer.SEPARATOR && current.Val != "," {
				panic("arguments must be separated by commas. Got: " + fmt.Sprintf("%+v", current))
			}

			p.i++
			current = p.input[p.i]
		}

		name := p.lookAhead(0)
		if name.Type != lexer.IDENTIFIER {
			panic("function arguments: variable name must be identifier. Got: " + fmt.Sprintf("%+v", name))
		}
		p.i++

		argType, err := p.parseOneType()
		if err != nil {
			panic(err)
		}
		p.i++

		res = append(res, &NameNode{
			Name: name.Val,
			Type: argType,
		})

		i++
	}
}

func (p *parser) parseOneType() (TypeNode, error) {
	current := p.lookAhead(0)

	isVariadic := false
	if current.Type == lexer.OPERATOR && current.Val == "..." {
		isVariadic = true

		p.i++
		current = p.lookAhead(0)
	}

	// pointer types
	if current.Type == lexer.OPERATOR && current.Val == "*" {
		p.i++
		valType, err := p.parseOneType()
		if err != nil {
			panic(err)
		}
		return &PointerTypeNode{
			ValueType:  valType,
			IsVariadic: isVariadic,
		}, nil
	}

	// struct parsing
	if current.Type == lexer.KEYWORD && current.Val == "struct" {
		p.i++

		res := &StructTypeNode{
			Types:      make([]TypeNode, 0),
			Names:      make(map[string]int),
			IsVariadic: isVariadic,
		}

		current = p.lookAhead(0)
		if current.Type != lexer.SEPARATOR || current.Val != "{" {
			panic("struct must be followed by {")
		}
		p.i++

		for {
			itemName := p.lookAhead(0)

			// Ignore EOL
			if itemName.Type == lexer.EOL {
				p.i++
				continue
			}

			// Stop at }
			if itemName.Type == lexer.SEPARATOR && itemName.Val == "}" {
				break
			}

			if itemName.Type != lexer.IDENTIFIER {
				panic("expected IDENTIFIER in struct{}, got " + fmt.Sprintf("%+v", itemName))
			}
			p.i++

			itemType, err := p.parseOneType()
			if err != nil {
				panic("expected TYPE in struct{}, got: " + err.Error())
			}
			p.i++

			res.Types = append(res.Types, itemType)
			res.Names[itemName.Val] = len(res.Types) - 1

			current = p.lookAhead(0)
		}

		return res, nil
	}

	if current.Type == lexer.KEYWORD && current.Val == "interface" {
		p.i++
		p.expect(p.lookAhead(0), lexer.Item{Type: lexer.SEPARATOR, Val: "{"})
		p.i++

		ifaceType := &InterfaceTypeNode{
			IsVariadic: isVariadic,
		}

		// Parse methods if set
		for {
			current := p.lookAhead(0)
			if current.Type == lexer.EOL {
				p.i++
				continue
			}
			if current.Type == lexer.SEPARATOR && current.Val == "}" {
				break
			}

			// Expect method name
			p.expect(current, lexer.Item{Type: lexer.IDENTIFIER})

			methodName := current.Val
			methodDef := InterfaceMethod{}

			p.i++
			p.expect(p.lookAhead(0), lexer.Item{Type: lexer.SEPARATOR, Val: "("})

			// Check if the method takes any arguments
			for {
				p.i++
				current = p.lookAhead(0)
				if current.Type == lexer.SEPARATOR && current.Val == ")" {
					p.i++
					break
				}

				current = p.lookAhead(0)
				if current.Type == lexer.SEPARATOR && current.Val == "," {
					continue
				}

				argumentType, err := p.parseOneType()
				if err != nil {
					panic(err)
				}

				methodDef.ArgumentTypes = append(methodDef.ArgumentTypes, argumentType)
			}

			// Function return types
			for {
				current = p.lookAhead(0)
				if current.Type == lexer.EOL {
					p.i++
					break
				}

				returnType, err := p.parseOneType()
				if err != nil {
					panic(err)
				}

				methodDef.ReturnTypes = append(methodDef.ReturnTypes, returnType)
				p.i++
			}

			if ifaceType.Methods == nil {
				ifaceType.Methods = make(map[string]InterfaceMethod)
			}
			ifaceType.Methods[methodName] = methodDef
		}

		return ifaceType, nil
	}

	if current.Type == lexer.IDENTIFIER {
		return &SingleTypeNode{
			TypeName:   current.Val,
			IsVariadic: isVariadic,
		}, nil
	}

	// Array parsing
	if current.Type == lexer.OPERATOR && current.Val == "[" {
		arrayLenght := p.lookAhead(1)

		// Slice parsing
		if arrayLenght.Type == lexer.OPERATOR && arrayLenght.Val == "]" {
			p.i += 2

			sliceItemType, err := p.parseOneType()
			if err != nil {
				return nil, errors.New("arrayParse failed: " + err.Error())
			}

			return &SliceTypeNode{
				ItemType:   sliceItemType,
				IsVariadic: isVariadic,
			}, nil
		}

		if arrayLenght.Type != lexer.NUMBER {
			return nil, errors.New("parseArray failed: Expected number or ] after [")
		}
		arrayLengthInt, err := strconv.Atoi(arrayLenght.Val)
		if err != nil {
			return nil, err
		}

		expectEndBracket := p.lookAhead(2)
		if expectEndBracket.Type != lexer.OPERATOR || expectEndBracket.Val != "]" {
			return nil, errors.New("parseArray failed: Expected ] in array type")
		}

		p.i += 3

		arrayItemType, err := p.parseOneType()
		if err != nil {
			return nil, errors.New("arrayParse failed: " + err.Error())
		}

		return &ArrayTypeNode{
			ItemType:   arrayItemType,
			Len:        int64(arrayLengthInt),
			IsVariadic: isVariadic,
		}, nil
	}

	// Func type parsing
	if current.Type == lexer.KEYWORD && current.Val == "func" {
		p.i++

		expectOpenParen := p.lookAhead(0)
		if expectOpenParen.Type != lexer.SEPARATOR && expectOpenParen.Val != "(" {
			return nil, errors.New("parse func failed, expected ( after func")
		}
		p.i++

		fn := &FuncTypeNode{}

		multiTypeParse := func() ([]TypeNode, error) {
			var typeList []TypeNode

			for {
				checkIfEndParen := p.lookAhead(0)
				if checkIfEndParen.Type == lexer.SEPARATOR && checkIfEndParen.Val == ")" {
					break
				}

				argType, err := p.parseOneType()
				if err != nil {
					return nil, err
				}

				typeList = append(typeList, argType)

				p.i++

				expectCommaOrEndParen := p.lookAhead(0)
				if expectCommaOrEndParen.Type == lexer.SEPARATOR && expectCommaOrEndParen.Val == "," {
					p.i++
					continue
				}

				if expectCommaOrEndParen.Type == lexer.SEPARATOR && expectCommaOrEndParen.Val == ")" {
					continue
				}

				return nil, errors.New("expected ) or , in func arg parsing")
			}

			return typeList, nil
		}

		// List of arguments
		var err error
		fn.ArgTypes, err = multiTypeParse()
		if err != nil {
			return nil, fmt.Errorf("unable to parse func arguments: %s", err)
		}

		// Return type parsing
		// Possible formats:
		// - Nothing
		// - T
		// - (T1, T2, ... )

		checkIfParenOrType := p.lookAhead(1)

		// Multiple types
		if checkIfParenOrType.Type == lexer.SEPARATOR && checkIfParenOrType.Val == "(" {
			p.i++
			p.i++

			fn.RetTypes, err = multiTypeParse()
			if err != nil {
				return nil, fmt.Errorf("unable to parse func return values: %s", err)
			}

			return fn, nil
		}

		// Single types
		if checkIfParenOrType.Type == lexer.IDENTIFIER {
			p.i++

			t, err := p.parseOneType()
			if err != nil {
				return nil, err
			}

			fn.RetTypes = []TypeNode{t}
			return fn, nil
		}

		return fn, nil
	}

	return nil, errors.New("parseOneType failed: " + fmt.Sprintf("%+v", current))
}

// panics if check fails
func (p *parser) expect(input lexer.Item, expected lexer.Item) {
	if expected.Type != input.Type {
		panic(fmt.Sprintf("Expected %+v got %+v", expected, input))
	}

	if expected.Val != "" && expected.Val != input.Val {
		panic(fmt.Sprintf("Expected %+v got %+v", expected, input))
	}
}
