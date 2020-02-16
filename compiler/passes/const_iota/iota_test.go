package const_iota

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zegl/tre/compiler/lexer"
	"github.com/zegl/tre/compiler/parser"
)

func TestIota(t *testing.T) {
	// Run input code through the lexer. A list of tokens is returned.
	lexed := lexer.Lex(`
package main

const (
	a = iota
	b = iota
	c = iota
)

`)

	parsed := parser.Parse(lexed, false)
	res := Iota(parsed)

	expected := &parser.FileNode{
		Instructions: []parser.Node{
			&parser.DeclarePackageNode{PackageName: "main"},
			&parser.AllocGroup{
				Allocs: []*parser.AllocNode{
					&parser.AllocNode{
						Name:    []string{"a"},
						Val:     []parser.Node{&parser.ConstantNode{Type: parser.NUMBER, Value: 0}},
						IsConst: true,
					},
					&parser.AllocNode{
						Name:    []string{"b"},
						Val:     []parser.Node{&parser.ConstantNode{Type: parser.NUMBER, Value: 1}},
						IsConst: true,
					},
					&parser.AllocNode{
						Name:    []string{"c"},
						Val:     []parser.Node{&parser.ConstantNode{Type: parser.NUMBER, Value: 2}},
						IsConst: true,
					},
				},
			},
		},
	}

	assert.Equal(t, expected, res)
}

func TestIotaShortSyntax(t *testing.T) {
	// Run input code through the lexer. A list of tokens is returned.
	lexed := lexer.Lex(`
package main

const (
	a = iota
	b
	c
)

`)

	parsed := parser.Parse(lexed, false)
	res := Iota(parsed)

	expected := &parser.FileNode{
		Instructions: []parser.Node{
			&parser.DeclarePackageNode{PackageName: "main"},
			&parser.AllocGroup{
				Allocs: []*parser.AllocNode{
					&parser.AllocNode{
						Name:    []string{"a"},
						Val:     []parser.Node{&parser.ConstantNode{Type: parser.NUMBER, Value: 0}},
						IsConst: true,
					},
					&parser.AllocNode{
						Name:    []string{"b"},
						Val:     []parser.Node{&parser.ConstantNode{Type: parser.NUMBER, Value: 1}},
						IsConst: true,
					},
					&parser.AllocNode{
						Name:    []string{"c"},
						Val:     []parser.Node{&parser.ConstantNode{Type: parser.NUMBER, Value: 2}},
						IsConst: true,
					},
				},
			},
		},
	}

	assert.Equal(t, expected, res)
}
