package parser

import (
	"github.com/zegl/tre/lexer"
	"log"
	"strconv"
)

type Node interface {
}

type CallNode struct {
	Function  string
	Arguments []Node
}

type BlockNode struct {
	Instructions []Node
}

type Operator uint8

const (
	OP_ADD Operator = iota
	OP_SUB
	OP_DIV
	OP_MUL
)

var opsCharToOp = map[string]Operator{
	"+": OP_ADD,
	"-": OP_SUB,
	"/": OP_DIV,
	"*": OP_MUL,
}

type OperatorNode struct {
	Operator Operator
	Left     Node
	Right    Node
}

type DataType uint8

const (
	STRING DataType = iota
	NUMBER
)

type ConstantNode struct {
	Type     DataType
	Value    int64
	ValueStr string
}

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
	}

	log.Panicf("unable to handle default: %+v", current)
	panic("")
}

func (p *parser) aheadParse(input Node) Node {
	next := p.lookAhead(1)

	if next.Type == lexer.OPERATOR {
		p.i += 2
		return OperatorNode{
			Operator: opsCharToOp[next.Val],
			Left:     input,
			Right:    p.parseOne(),
		}
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

		res = append(res, p.parseOne())
		p.i++
	}
}
