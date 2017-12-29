package parser

import (
	"github.com/zegl/tre/lexer"
	"log"
	"strconv"
	"fmt"
)

type parser struct {
	i     int
	input []lexer.Item
}

func Parse(input []lexer.Item) BlockNode {
	p := &parser{
		i:     0,
		input: input,
	}

	return BlockNode{
		Instructions: p.parseUntil(lexer.Item{Type: lexer.EOF}),
	}
}

func (p *parser) parseOne() Node {
	current := p.input[p.i]

	switch current.Type {
	case lexer.IDENTIFIER:
		next := p.lookAhead(1)
		if next.Type == lexer.SEPARATOR && next.Val == "(" {
			p.i += 2 // identifier and left paren
			return p.aheadParse(CallNode{
				Function:  current.Val,
				Arguments: p.parseUntil(lexer.Item{Type: lexer.SEPARATOR, Val: ")"}),
			})
		}

		return p.aheadParse(NameNode{
			Name: current.Val,
		})
		break

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

	case lexer.STRING:
		return p.aheadParse(ConstantNode{
			Type:     STRING,
			ValueStr: current.Val,
		})
		break

	case lexer.KEYWORD:
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
				Name:      name.Val,
				Arguments: arguments,
				ReturnValues: retTypesNodeNames,
				Body:      p.parseUntil(lexer.Item{Type: lexer.SEPARATOR, Val: "}"}),
			})
		}

		if current.Val == "return" {
			p.i++
			return p.aheadParse(ReturnNode{
				Val: p.parseOne(),
			})
		}
	}

	log.Panicf("unable to handle default: %+v", current)
	panic("")
}

func (p *parser) aheadParse(input Node) Node {
	next := p.lookAhead(1)

	if next.Type == lexer.OPERATOR {
		p.i += 2

		if next.Val == ":=" {
			if nameNode, ok := input.(NameNode); ok {
				return AllocNode{
					Name: nameNode.Name,
					Val:  p.parseOne(),
				}
			} else {
				panic(":= can only be used after a name")
			}
		}

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

	return input
}

func (p *parser) lookAhead(steps int) lexer.Item {
	return p.input[p.i+steps]
}

func (p *parser) parseUntil(until lexer.Item) []Node {
	var res []Node

	for {
		current := p.input[p.i]
		if current.Type == until.Type && current.Val == until.Val {
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
