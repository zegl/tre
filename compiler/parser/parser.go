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

func Parse(input []lexer.Item, debug bool) BlockNode {
	p := &parser{
		i:     0,
		input: input,
		debug: debug,
	}

	return BlockNode{
		Instructions: p.parseUntil(lexer.Item{Type: lexer.EOF}),
	}
}

func (p *parser) parseOne() Node {
	current := p.input[p.i]

	if p.debug {
		fmt.Printf("parseOne: %d - %+v\n", p.i, current)
	}

	switch current.Type {

	case lexer.EOF:
		panic("unexpected EOF")

	// IDENTIFIERS are converted to either:
	// - a CallNode if followed by an opening parenthesis (a function call), or
	// - a NodeName (variables)
	case lexer.IDENTIFIER:
		return p.aheadParse(NameNode{
			Name: current.Val,
		})

		// NUMBER always returns a ConstantNode
		// Convert string representation to int64
	case lexer.NUMBER:
		val, err := strconv.ParseInt(current.Val, 10, 64)
		if err != nil {
			panic(err)
		}

		return p.aheadParse(ConstantNode{
			Type:  NUMBER,
			Value: val,
		})

		// STRING is always a ConstantNode, the value is not modified
	case lexer.STRING:
		return p.aheadParse(ConstantNode{
			Type:     STRING,
			ValueStr: current.Val,
		})

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

			cond, ok := condNodes[0].(OperatorNode)
			if !ok {
				panic("node in if-condition must be OperatorNode")
			}

			p.i++

			return p.aheadParse(ConditionNode{
				Cond: cond,
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

				if singleTypeNode, ok := methodOnType.(SingleTypeNode); ok {
					defineFunc.MethodOnType = singleTypeNode
				} else {
					panic("could not find type in method defitition")
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
				retType, err := p.parseOneType()
				if err != nil {
					panic(err)
				}
				retTypesNodeNames = append(retTypesNodeNames, NameNode{
					Type: retType.(SingleTypeNode),
				})
				p.i++
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
			return p.aheadParse(ReturnNode{
				Val: p.parseOne(),
			})
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

			return p.aheadParse(DefineTypeNode{
				Name: name.Val,
				Type: typeType,
			})
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
			if next.Type != lexer.IDENTIFIER {
				panic(fmt.Sprintf("Expected IDENTFIER after . Got: %+v", next))
			}

			p.i++

			return p.aheadParse(StructLoadElementNode{
				Struct:      input,
				ElementName: next.Val,
			})
		}

		if next.Val == ":=" || next.Val == "=" {
			p.i += 2

			if nameNode, ok := input.(NameNode); ok {
				if next.Val == ":=" {
					return AllocNode{
						Name: nameNode.Name,
						Val:  p.parseOne(),
					}
				} else {
					return AssignNode{
						Name: nameNode.Name,
						Val:  p.parseOne(),
					}
				}
			}

			if next.Val == "=" {
				if loadNode, ok := input.(StructLoadElementNode); ok {
					return AssignNode{
						Target: loadNode,
						Val:    p.parseOne(),
					}
				}

				if arrayNode, ok := input.(LoadArrayElement); ok {
					return AssignNode{
						Target: arrayNode,
						Val:    p.parseOne(),
					}
				}
			}

			panic(fmt.Sprintf("%s can only be used after a name. Got: %+v", next.Val, input))
		}

		// Array slicing
		if next.Val == "[" {
			p.i += 2

			index := p.parseOne()

			var res Node

			checkIfColon := p.lookAhead(1)
			if checkIfColon.Type == lexer.OPERATOR && checkIfColon.Val == ":" {
				p.i += 2
				res = SliceArrayNode{
					Val:    input,
					Start:  index,
					HasEnd: true,
					End:    p.parseOne(),
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
				Right:    p.parseOne(),
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
				Type: current.Val,
				Val:  val[0],
			})
		}

		return p.aheadParse(CallNode{
			Function:  input,
			Arguments: p.parseUntil(lexer.Item{Type: lexer.SEPARATOR, Val: ")"}),
		})
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

		res = append(res, p.parseOne())
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
			Type: argType.(SingleTypeNode),
		})

		i++
	}
}

func (p *parser) parseOneType() (TypeNode, error) {
	current := p.lookAhead(0)

	// struct parsing
	if current.Type == lexer.KEYWORD && current.Val == "struct" {
		p.i++

		res := StructTypeNode{
			Types: make([]TypeNode, 0),
			Names: make(map[string]int),
		}

		current = p.lookAhead(0)
		if current.Type != lexer.SEPARATOR || current.Val != "{" {
			panic("struct must be followed by {")
		}
		p.i++

		for current.Type != lexer.SEPARATOR || current.Val != "}" {
			itemName := p.lookAhead(0)
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

	if current.Type == lexer.IDENTIFIER {
		return SingleTypeNode{
			TypeName: current.Val,
		}, nil
	}

	// Array parsing
	if current.Type == lexer.OPERATOR && current.Val == "[" {
		arrayLenght := p.lookAhead(1)
		if arrayLenght.Type != lexer.NUMBER {
			return nil, errors.New("parseArray failed: Expected number after [")
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
			ItemType: arrayItemType,
			Len:      int64(arrayLengthInt),
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
