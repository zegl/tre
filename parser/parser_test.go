package parser

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/zegl/tre/lexer"
)

func TestCall(t *testing.T) {
	input := []lexer.Item{
		{lexer.IDENTIFIER, "printf"},
		{lexer.SEPARATOR, "("},
		{lexer.NUMBER, "1"},
		{lexer.SEPARATOR, ")"},
		{lexer.EOF, ""},
	}

	expected := BlockNode{
		Instructions: []Node{
			CallNode{
				Function:  "printf",
				Arguments: []Node{ConstantNode{Type: NUMBER, Value: 1}},
			},
		},
	}

	assert.Equal(t, expected, Parse(input))
}

func TestAdd(t *testing.T) {
	input := []lexer.Item{
		{lexer.NUMBER, "1"},
		{lexer.OPERATOR, "+"},
		{lexer.NUMBER, "2"},
		{lexer.EOF, ""},
	}

	expected := BlockNode{
		Instructions: []Node{
			OperatorNode{
				Operator: OP_ADD,
				Left: ConstantNode{
					Type:  NUMBER,
					Value: 1,
				},
				Right: ConstantNode{
					Type:  NUMBER,
					Value: 2,
				},
			},
		},
	}

	assert.Equal(t, expected, Parse(input))
}
