package parser

import (
	"github.com/zegl/tre/lexer"
	"log"
	"strconv"
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

		panic("unable to handle identifier")
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
	}

	log.Panicf("unable to handle default: %+v", current)
	panic("")
}

func (p *parser) aheadParse(input Node) Node {
	next := p.lookAhead(1)

	if next.Type == lexer.OPERATOR {
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
