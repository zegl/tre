package parser

import (
	"github.com/zegl/tre/lexer"
	"log"
	"strconv"
	"fmt"
)

type Node interface {
}

type CallNode struct {
	Function  string
	Arguments []Node
}

func (cn CallNode) String() string {
	return fmt.Sprintf("CallNode: %s(%+v)", cn.Function, cn.Arguments)
}

type BlockNode struct {
	Instructions []Node
}

func (bn BlockNode) String() string {
	return fmt.Sprintf("BlockNode: %+v", bn.Instructions)
}

type Operator uint8

const (
	OP_ADD Operator = '+'
	OP_SUB          = '-'
	OP_DIV          = '/'
	OP_MUL          = '*'
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

func (on OperatorNode) String() string {
	return fmt.Sprintf("(%v %s %v)", on.Left, string(on.Operator), on.Right)
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

func (cn ConstantNode) String() string {
	return fmt.Sprintf("%d", cn.Value)
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

	case lexer.STRING:
		return p.aheadParse(ConstantNode{
			Type:     STRING,
			ValueStr: current.Val,
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
