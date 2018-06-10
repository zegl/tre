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

	debug bool
}

func Parse(input []lexer.Item, debug bool) FileNode {
	p := &parser{
		i:     0,
		input: input,
		debug: debug,
	}

	return FileNode{
		Instructions: p.parseUntil(lexer.Item{Type: lexer.EOF}),
	}
}

func (p *parser) parseOne(withAheadParse bool) (res Node) {
	current := p.input[p.i]

	if p.debug {
		fmt.Printf("parseOne: %d - %+v\n", p.i, current)
	}

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
		res = NameNode{Name: current.Val}
		if withAheadParse {
			res = p.aheadParse(res)
		}
		return

		// NUMBER always returns a ConstantNode
		// Convert string representation to int64
	case lexer.NUMBER:
		val, err := strconv.ParseInt(current.Val, 10, 64)
		if err != nil {
			panic(err)
		}

		res = ConstantNode{
			Type:  NUMBER,
			Value: val,
		}
		if withAheadParse {
			res = p.aheadParse(res)
		}
		return

		// STRING is always a ConstantNode, the value is not modified
	case lexer.STRING:
		res = ConstantNode{
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
			res = GetReferenceNode{Item: p.parseOne(true)}
			if withAheadParse {
				res = p.aheadParse(res)
			}
			return
		}
		if current.Val == "*" {
			p.i++
			res = DereferenceNode{Item: p.parseOne(false)}
			if withAheadParse {
				res = p.aheadParse(res)
			}
			return
		}

		if current.Val == "!" {
			p.i++
			res = NegateNode{Item: p.parseOne(false)}
			return
		}

		// Slice initalization
		if current.Val == "[" {
			next := p.lookAhead(1)
			if next.Type != lexer.OPERATOR || next.Val != "]" {
				panic("expected ] after [")
			}

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

			items := p.parseUntil(lexer.Item{Type: lexer.SEPARATOR, Val: "}"})

			res = InitializeSliceNode{
				Type:  sliceItemType,
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
			p.i++
			condNodes := p.parseUntil(lexer.Item{Type: lexer.SEPARATOR, Val: "{"})

			if len(condNodes) != 1 {
				panic("could not parse if-condition")
			}

			var opNode OperatorNode

			if cond, ok := condNodes[0].(OperatorNode); ok {
				opNode = cond
			} else {
				// Add implicit == true
				opNode = OperatorNode{
					Left: condNodes[0],
					Right: ConstantNode{
						Type:  BOOL,
						Value: 1,
					},
					Operator: OP_EQ,
				}
			}

			p.i++

			return p.aheadParse(ConditionNode{
				Cond: opNode,
				True: p.parseUntil(lexer.Item{Type: lexer.SEPARATOR, Val: "}"}),
			})
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
		if current.Val == "func" {
			defineFunc := DefineFuncNode{}
			p.i++

			// Check if next is IDENTIFIER (named function), or an opening parenthesis (method).
			checkIfOpeningParen := p.lookAhead(0)

			// Method Parsning
			if checkIfOpeningParen.Type == lexer.SEPARATOR && checkIfOpeningParen.Val == "(" {
				p.i++

				expectIdentifier := p.lookAhead(0)
				if expectIdentifier.Type != lexer.IDENTIFIER {
					panic("could not find type identifier in method definition")
				}

				defineFunc.IsMethod = true
				defineFunc.InstanceName = expectIdentifier.Val

				p.i++

				methodOnType, err := p.parseOneType()
				if err != nil {
					panic(err)
				}

				if pointerSingleTypeNode, ok := methodOnType.(PointerTypeNode); ok {
					defineFunc.IsPointerReceiver = true
					methodOnType = pointerSingleTypeNode.ValueType
				}

				if singleTypeNode, ok := methodOnType.(SingleTypeNode); ok {
					defineFunc.MethodOnType = singleTypeNode
				} else {
					panic(fmt.Sprintf("could not find type in method defitition: %T", methodOnType))
				}

				p.i++

				// Expect closing paren
				expectCloseParen := p.lookAhead(0)
				if expectCloseParen.Type != lexer.SEPARATOR || expectCloseParen.Val != ")" {
					panic("expected ) after method type in method definition")
				}

				p.i++
			}

			name := p.lookAhead(0)
			if name.Type != lexer.IDENTIFIER {
				panic("func must be followed by IDENTIFIER. Got " + name.Val)
			}
			defineFunc.Name = name.Val

			p.i++

			openParen := p.lookAhead(0)
			if openParen.Type != lexer.SEPARATOR || openParen.Val != "(" {
				panic("func identifier must be followed by (. Got " + openParen.Val)
			}

			p.i++

			// Parse argument list
			defineFunc.Arguments = p.parseFunctionArguments()

			// Parse return types
			var retTypesNodeNames []NameNode

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
					retType, err := p.parseOneType()
					if err != nil {
						panic(err)
					}
					retTypesNodeNames = append(retTypesNodeNames, NameNode{
						Type: retType,
					})
					p.i++

					if !allowMultiRetVals {
						break
					}

					// Require comma or end parenthesis
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

			res = ReturnNode{Vals: retVals}
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

			res = DefineTypeNode{
				Name: name.Val,
				Type: typeType,
			}
			if withAheadParse {
				res = p.aheadParse(res)
			}

			// Register that this type exists
			// TODO: Make it context sensitive (eg package level types, types in functions etc)
			types[name.Val] = struct{}{}

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

			return AllocNode{
				Name: name.Val,
				Val:  tp,
			}
		}

		if current.Val == "package" {
			packageName := p.lookAhead(1)

			if packageName.Type != lexer.IDENTIFIER {
				panic("package must be followed by a IDENTIFIER")
			}

			p.i += 1

			return DeclarePackageNode{
				PackageName: packageName.Val,
			}
		}

		if current.Val == "for" {
			return p.parseFor()
		}

		if current.Val == "break" {
			return BreakNode{}
		}

		if current.Val == "continue" {
			return ContinueNode{}
		}

		if current.Val == "import" {
			return p.parseImport()
		}

		if current.Val == "true" || current.Val == "false" {
			var v int64 = 0
			if current.Val == "true" {
				v = 1
			}

			return ConstantNode{
				Type:  BOOL,
				Value: v,
			}
		}
	}

	p.printInput()
	log.Panicf("unable to handle default: %d - %+v", p.i, current)
	panic("")
}

func (p *parser) aheadParse(input Node) Node {

	next := p.lookAhead(1)

	if p.debug {
		fmt.Printf("aheadParse: %+v - %+v\n", next, input)
	}

	if next.Type == lexer.OPERATOR {
		if next.Val == "." {
			p.i++

			next = p.lookAhead(1)
			if next.Type == lexer.IDENTIFIER {
				p.i++
				return p.aheadParse(StructLoadElementNode{
					Struct:      input,
					ElementName: next.Val,
				})
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

				return p.aheadParse(TypeCastInterfaceNode{
					Item: input,
					Type: castToType,
				})
			}

			panic(fmt.Sprintf("Expected IDENTFIER or ( after . Got: %+v", next))
		}

		if next.Val == ":=" || next.Val == "=" {
			p.i += 2

			if nameNode, ok := input.(NameNode); ok {
				if next.Val == ":=" {
					return AllocNode{
						Name: nameNode.Name,
						Val:  p.parseOne(true),
					}
				}
				return AssignNode{
					Name: nameNode.Name,
					Val:  p.parseOne(true),
				}
			}

			if next.Val == "=" {
				if loadNode, ok := input.(StructLoadElementNode); ok {
					return AssignNode{
						Target: loadNode,
						Val:    p.parseOne(true),
					}
				}

				if arrayNode, ok := input.(LoadArrayElement); ok {
					return AssignNode{
						Target: arrayNode,
						Val:    p.parseOne(true),
					}
				}

				if dereferenceNode, ok := input.(DereferenceNode); ok {
					return AssignNode{
						Target: dereferenceNode,
						Val:    p.parseOne(true),
					}
				}
			}

			panic(fmt.Sprintf("%s can only be used after a name. Got: %+v", next.Val, input))
		}

		if next.Val == "..." {
			p.i++
			return DeVariadicSliceNode{
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
				res = SliceArrayNode{
					Val:    input,
					Start:  index,
					HasEnd: true,
					End:    p.parseOne(true),
				}
			} else {
				res = LoadArrayElement{
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

		if _, ok := opsCharToOp[next.Val]; ok {
			p.i += 2
			res := OperatorNode{
				Operator: opsCharToOp[next.Val],
				Left:     input,
				Right:    p.parseOne(true),
			}

			// Sort infix operations if necessary (eg: apply OP_MUL before OP_ADD)
			if right, ok := res.Right.(OperatorNode); ok {
				return sortInfix(res, right)
			}

			return p.aheadParse(res)
		}
	}

	if next.Type == lexer.SEPARATOR && next.Val == "(" {
		current := p.lookAhead(0)

		p.i += 2 // identifier and left paren

		if _, ok := types[current.Val]; ok {
			val := p.parseUntil(lexer.Item{Type: lexer.SEPARATOR, Val: ")"})
			if len(val) != 1 {
				panic("type conversion must take only one argument")
			}
			return p.aheadParse(TypeCastNode{
				Type: SingleTypeNode{
					TypeName: current.Val,
				},
				Val: val[0],
			})
		}

		return p.aheadParse(CallNode{
			Function:  input,
			Arguments: p.parseUntil(lexer.Item{Type: lexer.SEPARATOR, Val: ")"}),
		})
	}

	// Initialize structs with values:
	//   Foo{Bar: 123}
	//   Foo{Bar: 123, Bax: hello(123)}
	if next.Type == lexer.SEPARATOR && next.Val == "{" {
		nameNode, isNamedNode := input.(NameNode)
		if isNamedNode {
			_, isType := types[nameNode.Name]
			if isType {
				inputType := SingleTypeNode{
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

					items[key.Val] = p.parseOne(true)

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

				return p.aheadParse(InitializeStructNode{
					Type:  inputType,
					Items: items,
				})
			}
		}
	}

	if next.Type == lexer.SEPARATOR && next.Val == "," {
		if inputNamedNode, ok := input.(NameNode); ok {

			// This bit of parsing is specualtive
			//
			// The parser will restore it's position to this index in case it
			// turns out that we can not convert this into a MultiNameNode
			preIndex := p.i

			p.i++
			p.i++

			nextName := p.parseOne(true)

			if nextAlloc, ok := nextName.(AllocNode); ok {
				if len(nextAlloc.MultiNames.Names) == 0 {
					nextAlloc.MultiNames = MultiNameNode{
						Names: []NameNode{inputNamedNode, NameNode{Name: nextAlloc.Name}},
					}
					nextAlloc.Name = ""
				} else {
					// Add the current one to the beging of the list
					newMultiNames := []NameNode{inputNamedNode}
					newMultiNames = append(newMultiNames, nextAlloc.MultiNames.Names...)

					nextAlloc.MultiNames = MultiNameNode{Names: newMultiNames}
				}

				return p.aheadParse(nextAlloc)
			}

			// A MultiNameNode could not be created
			// Reset the parsing index
			p.i = preIndex
		}
	}

	return input
}

func (p *parser) lookAhead(steps int) lexer.Item {
	return p.input[p.i+steps]
}

func (p *parser) parseUntil(until lexer.Item) []Node {
	var res []Node

	if p.debug {
		fmt.Printf("parseUntil: %+v\n", until)
	}

	for {
		if p.debug {
			fmt.Printf("in parseUntil: %+v\n", until)
		}

		current := p.input[p.i]
		if current.Type == until.Type && current.Val == until.Val {
			if p.debug {
				fmt.Printf("parseUntil: %+v done\n", until)
			}
			return res
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

func (p *parser) parseFunctionArguments() []NameNode {
	var res []NameNode
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

		res = append(res, NameNode{
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
		return PointerTypeNode{
			ValueType:  valType,
			IsVariadic: isVariadic,
		}, nil
	}

	// struct parsing
	if current.Type == lexer.KEYWORD && current.Val == "struct" {
		p.i++

		res := StructTypeNode{
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

		ifaceType := InterfaceTypeNode{
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
		return SingleTypeNode{
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

			return SliceTypeNode{
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

		return ArrayTypeNode{
			ItemType:   arrayItemType,
			Len:        int64(arrayLengthInt),
			IsVariadic: isVariadic,
		}, nil
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
