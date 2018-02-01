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
	// IDENTIFIERS are converted to either:
	// - a CallNode if followed by an opening parenthesis (a function call), or
	// - a NodeName (variables)
	case lexer.IDENTIFIER:
		next := p.lookAhead(1)
		if next.Type == lexer.SEPARATOR && next.Val == "(" {
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
				Function:  current.Val,
				Arguments: p.parseUntil(lexer.Item{Type: lexer.SEPARATOR, Val: ")"}),
			})
		}

		return p.aheadParse(NameNode{
			Name: current.Val,
		})
		break

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
		break

		// STRING is always a ConstantNode, the value is not modified
	case lexer.STRING:
		return p.aheadParse(ConstantNode{
			Type:     STRING,
			ValueStr: current.Val,
		})
		break

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
			name := p.lookAhead(1)
			if name.Type != lexer.IDENTIFIER {
				panic("func must be followed by IDENTIFIER. Got " + name.Val)
			}

			p.i++

			openParen := p.lookAhead(1)
			if openParen.Type != lexer.SEPARATOR || openParen.Val != "(" {
				panic("func identifier must be followed by (. Got " + openParen.Val)
			}

			p.i++
			p.i++

			arguments := p.parseFunctionArguments()

			retTypes := p.parseUntil(lexer.Item{Type: lexer.SEPARATOR, Val: "{"})
			// convert return types
			var retTypesNodeNames []NameNode
			for _, r := range retTypes {
				rr := r.(NameNode)
				retTypesNodeNames = append(retTypesNodeNames, NameNode{
					Type: rr.Name,
				})
			}

			openBracket := p.lookAhead(0)
			if openBracket.Type != lexer.SEPARATOR || openBracket.Val != "{" {
				panic("func arguments must be followed by {. Got " + openBracket.Val)
			}

			p.i++

			return p.aheadParse(DefineFuncNode{
				Name:         name.Val,
				Arguments:    arguments,
				ReturnValues: retTypesNodeNames,
				Body:         p.parseUntil(lexer.Item{Type: lexer.SEPARATOR, Val: "}"}),
			})
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
		if current.Val == "new" {
			p.i++

			tp, err := p.parseOneType()
			if err != nil {
				panic(err)
			}

			return tp
		}
	}

	p.printInput()
	log.Panicf("unable to handle default: %d - %+v", p.i, current)
	panic("")
}

func (p *parser) aheadParse(input Node) Node {
	next := p.lookAhead(1)

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
			}

			panic(fmt.Sprintf("%s can only be used after a name. Got: %+v", next.Val, input))
		}

		// Array slicing
		if next.Val == "[" {
			p.i += 2

			res := SliceArrayNode{
				Val:   input,
				Start: p.parseOne(),
			}

			checkIfColon := p.lookAhead(1)
			if checkIfColon.Type == lexer.OPERATOR && checkIfColon.Val == ":" {
				p.i += 2
				res.HasEnd = true
				res.End = p.parseOne()
			}

			p.i++

			expectEndBracket := p.lookAhead(0)
			if expectEndBracket.Type == lexer.OPERATOR && expectEndBracket.Val == "]" {
				return res
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

			return res
		}

		log.Printf("parse ahead do nothing on: %+v", next)
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

		typeNode := p.lookAhead(1)
		if typeNode.Type != lexer.IDENTIFIER {
			panic("function arguments: variable type must be identifier. Got: " + fmt.Sprintf("%+v", typeNode))
		}

		res = append(res, NameNode{
			Name: name.Val,
			Type: typeNode.Val,
		})

		p.i += 2
		i++
	}
}

func (p *parser) parseOneType() (TypeNode, error) {
	current := p.lookAhead(0)

	// struct parsing
	if current.Type == lexer.KEYWORD && current.Val == "struct" {
		p.i++

		res := &StructTypeNode{
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
		return &SingleTypeNode{
			TypeName: current.Val,
		}, nil
	}

	return nil, errors.New("parseOneType failed: " + fmt.Sprintf("%+v", current))
}
