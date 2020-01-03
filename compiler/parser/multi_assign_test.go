package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zegl/tre/compiler/lexer"
)

func TestMultiAssignVar(t *testing.T) {
	input := []lexer.Item{
		{Type: lexer.KEYWORD, Val: "package", Line: 1},
		{Type: lexer.IDENTIFIER, Val: "main", Line: 1},
		{Type: lexer.EOL, Val: "", Line: 1},
		{Type: lexer.EOL, Val: "", Line: 2},
		{Type: lexer.KEYWORD, Val: "func", Line: 3},
		{Type: lexer.IDENTIFIER, Val: "main", Line: 3},
		{Type: lexer.OPERATOR, Val: "(", Line: 3},
		{Type: lexer.OPERATOR, Val: ")", Line: 3},
		{Type: lexer.OPERATOR, Val: "{", Line: 3},
		{Type: lexer.EOL, Val: "", Line: 3},
		{Type: lexer.IDENTIFIER, Val: "a", Line: 4},
		{Type: lexer.OPERATOR, Val: ":=", Line: 4},
		{Type: lexer.NUMBER, Val: "100", Line: 4},
		{Type: lexer.EOL, Val: "", Line: 4},
		{Type: lexer.IDENTIFIER, Val: "b", Line: 5},
		{Type: lexer.OPERATOR, Val: ":=", Line: 5},
		{Type: lexer.NUMBER, Val: "200", Line: 5},
		{Type: lexer.EOL, Val: "", Line: 5},
		{Type: lexer.IDENTIFIER, Val: "a", Line: 6},
		{Type: lexer.OPERATOR, Val: ",", Line: 6},
		{Type: lexer.IDENTIFIER, Val: "b", Line: 6},
		{Type: lexer.OPERATOR, Val: "=", Line: 6},
		{Type: lexer.NUMBER, Val: "300", Line: 6},
		{Type: lexer.OPERATOR, Val: ",", Line: 6},
		{Type: lexer.NUMBER, Val: "400", Line: 6},
		{Type: lexer.EOL, Val: "", Line: 6},
		{Type: lexer.OPERATOR, Val: "}", Line: 7},
		{Type: lexer.EOL, Val: "", Line: 7},
		{Type: lexer.EOF, Val: "", Line: 0},
	}

	expected := FileNode{
		Instructions: []Node{
			&DeclarePackageNode{PackageName: "main"},
			&DefineFuncNode{Name: "main", IsNamed: true,
				Body: []Node{
					&AllocNode{Name: []string{"a"}, Val: []Node{&ConstantNode{Type: NUMBER, Value: 100}}},
					&AllocNode{Name: []string{"b"}, Val: []Node{&ConstantNode{Type: NUMBER, Value: 200}}},
					&AssignNode{
						Target: []Node{&NameNode{Name: "a"}, &NameNode{Name: "b"}},
						Val:    []Node{&ConstantNode{Type: NUMBER, Value: 300}, &ConstantNode{Type: NUMBER, Value: 400}},
					},
				},
			},
		},
	}

	assert.Equal(t, expected, Parse(input, false))
}
func TestMultiAllocVar(t *testing.T) {
	input := []lexer.Item{
		{Type: lexer.KEYWORD, Val: "package", Line: 1},
		{Type: lexer.IDENTIFIER, Val: "main", Line: 1},
		{Type: lexer.EOL, Val: "", Line: 1},
		{Type: lexer.EOL, Val: "", Line: 2},
		{Type: lexer.KEYWORD, Val: "func", Line: 3},
		{Type: lexer.IDENTIFIER, Val: "main", Line: 3},
		{Type: lexer.OPERATOR, Val: "(", Line: 3},
		{Type: lexer.OPERATOR, Val: ")", Line: 3},
		{Type: lexer.OPERATOR, Val: "{", Line: 3},
		{Type: lexer.EOL, Val: "", Line: 3},
		{Type: lexer.IDENTIFIER, Val: "a", Line: 6},
		{Type: lexer.OPERATOR, Val: ",", Line: 6},
		{Type: lexer.IDENTIFIER, Val: "b", Line: 6},
		{Type: lexer.OPERATOR, Val: ":=", Line: 6},
		{Type: lexer.NUMBER, Val: "300", Line: 6},
		{Type: lexer.OPERATOR, Val: ",", Line: 6},
		{Type: lexer.NUMBER, Val: "400", Line: 6},
		{Type: lexer.EOL, Val: "", Line: 6},
		{Type: lexer.OPERATOR, Val: "}", Line: 7},
		{Type: lexer.EOL, Val: "", Line: 7},
		{Type: lexer.EOF, Val: "", Line: 0},
	}

	expected := FileNode{
		Instructions: []Node{
			&DeclarePackageNode{PackageName: "main"},
			&DefineFuncNode{Name: "main", IsNamed: true,
				Body: []Node{
					&AllocNode{
						Name: []string{"a", "b"},
						Val:  []Node{&ConstantNode{Type: NUMBER, Value: 300}, &ConstantNode{Type: NUMBER, Value: 400}},
					},
				},
			},
		},
	}

	assert.Equal(t, expected, Parse(input, false))
}
