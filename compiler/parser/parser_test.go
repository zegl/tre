package parser

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/zegl/tre/compiler/lexer"
)

func TestCall(t *testing.T) {
	input := []lexer.Item{
		{Type: lexer.IDENTIFIER, Val: "printf"},
		{Type: lexer.SEPARATOR, Val: "("},
		{Type: lexer.NUMBER, Val: "1"},
		{Type: lexer.SEPARATOR, Val: ")"},
		{Type: lexer.EOF, Val: ""},
	}

	expected := BlockNode{
		Instructions: []Node{
			CallNode{
				Function:  "printf",
				Arguments: []Node{ConstantNode{Type: NUMBER, Value: 1}},
			},
		},
	}

	assert.Equal(t, expected, Parse(input, false))
}

func TestAdd(t *testing.T) {
	input := []lexer.Item{
		{Type: lexer.NUMBER, Val: "1"},
		{Type: lexer.OPERATOR, Val: "+"},
		{Type: lexer.NUMBER, Val: "2"},
		{Type: lexer.EOF, Val: ""},
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

	assert.Equal(t, expected, Parse(input, false))
}

func TestInfixPriority(t *testing.T) {
	input := []lexer.Item{
		{Type: lexer.NUMBER, Val: "1"},
		{Type: lexer.OPERATOR, Val: "+"},
		{Type: lexer.NUMBER, Val: "2"},
		{Type: lexer.OPERATOR, Val: "*"},
		{Type: lexer.NUMBER, Val: "3"},
		{Type: lexer.EOF, Val: ""},
	}

	expected := BlockNode{
		Instructions: []Node{
			OperatorNode{
				Operator: OP_ADD,
				Left: ConstantNode{
					Type:  NUMBER,
					Value: 1,
				},
				Right: OperatorNode{
					Operator: OP_MUL,
					Left: ConstantNode{
						Type:  NUMBER,
						Value: 2,
					},
					Right: ConstantNode{
						Type:  NUMBER,
						Value: 3,
					},
				},
			},
		},
	}

	assert.Equal(t, expected, Parse(input, false))
}

func TestInfixPriority2(t *testing.T) {
	input := []lexer.Item{
		{Type: lexer.NUMBER, Val: "1"},
		{Type: lexer.OPERATOR, Val: "*"},
		{Type: lexer.NUMBER, Val: "2"},
		{Type: lexer.OPERATOR, Val: "+"},
		{Type: lexer.NUMBER, Val: "3"},
		{Type: lexer.EOF, Val: ""},
	}

	expected := BlockNode{
		Instructions: []Node{
			OperatorNode{
				Operator: OP_ADD,
				Left: OperatorNode{
					Operator: OP_MUL,
					Left: ConstantNode{
						Type:  NUMBER,
						Value: 1,
					},
					Right: ConstantNode{
						Type:  NUMBER,
						Value: 2,
					},
				},
				Right: ConstantNode{
					Type:  NUMBER,
					Value: 3,
				},
			},
		},
	}

	assert.Equal(t, expected, Parse(input, false))
}
