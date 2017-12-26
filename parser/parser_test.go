package parser

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/zegl/tre/lexer"
)

func TestAdd(t *testing.T) {
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
