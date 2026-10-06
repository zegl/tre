package escape

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zegl/tre/compiler/lexer"
	"github.com/zegl/tre/compiler/parser"
)

func escapeTest(t *testing.T, input string, expected map[string]bool) {
	lexed := lexer.Lex(input)
	parsed := parser.Parse(lexed, false)
	parsed = Escape(parsed)

	var allocsChecked []string

	for _, ins := range parsed.Instructions {
		if defFuncNode, ok := ins.(*parser.DefineFuncNode); ok {
			for _, ins := range defFuncNode.Body {
				if allocNode, ok := ins.(*parser.AllocNode); ok {
					allocsChecked = append(allocsChecked, allocNode.Name[0])
					assert.Equal(t, expected[allocNode.Name[0]], allocNode.Escapes, allocNode.Name)
				}
			}
		}
	}

	assert.Equal(t, len(allocsChecked), len(expected))
}

func TestNoEscape(t *testing.T) {
	escapeTest(t, `package main

	func main() {
		a := 100
		b := 200
	}
`, map[string]bool{
		"a": false,
		"b": false,
	})
}

func TestEscapes(t *testing.T) {
	escapeTest(t, `package main

		func main() {
			a := 100
			b := 200
			return b
		}
	`, map[string]bool{
		"a": false,
		"b": true,
	})
}

func TestEscapesPointer(t *testing.T) {
	escapeTest(t, `package main

		func main() *int {
			a := 100
			b := 200
			return &b
		}
	`, map[string]bool{
		"a": false,
		"b": true,
	})
}

func TestEscapesStructPointer(t *testing.T) {
	escapeTest(t, `package main

		type mytype struct {
			a int
			b int
		}

		func main() *int {
			a := 100
			b := mytype{
				a: 100,
				b: 200,
			}
			return &b
		}
	`, map[string]bool{
		"a": false,
		"b": true,
	})
}

func TestEscapeNestedStruct(t *testing.T) {
	escapeTest(t, `package main

		type Bar struct {
			num int64
		}

		type Foo struct {
			num int64
			bar *Bar
		}

		func GetFooPtr() *Foo {
			f := Foo{
				num: 300,
				bar: &Bar{num: 400},
			}

			return &f
		}`,
		map[string]bool{
			"f": true,
		})
}

/*
TODO: Implement feature so that this case can pass
f can be stack allocated, but f.bar needs to allocqated on the heap
func TestNoEscapeNestedStruct(t *testing.T) {
	escapeTest(t, `package main

		type Bar struct {
			num int64
		}

		type Foo struct {
			num int64
			bar *Bar
		}

		func GetFooPtr() Foo {
			f := Foo{
				num: 300,
				bar: &Bar{num: 400},
			}

			return f
		}`,
		map[string]bool{
			"f": false,
		})
}
*/

type allocCollector map[string]bool

func (c allocCollector) Visit(node parser.Node) (parser.Node, parser.Visitor) {
	if allocNode, ok := node.(*parser.AllocNode); ok {
		for _, name := range allocNode.Name {
			c[name] = allocNode.Escapes
		}
	}
	return node, c
}

func nestedEscapeTest(t *testing.T, input string, expected map[string]bool) {
	parsed := Escape(parser.Parse(lexer.Lex(input), false))

	allocs := allocCollector{}
	parser.Walk(allocs, parsed)
	assert.Equal(t, expected, map[string]bool(allocs))
}

func TestEscapesInNestedBlock(t *testing.T) {
	nestedEscapeTest(t, `package main

		func main() *int {
			a := 100
			if a > 10 {
				b := 200
				return &b
			}
			for i := 0; i < 10; i++ {
				c := 300
				return &c
			}
			return &a
		}
	`, map[string]bool{
		"a": true,
		"b": true,
		"i": false,
		"c": true,
	})
}

func TestEscapesInFuncLiteral(t *testing.T) {
	nestedEscapeTest(t, `package main

		func main() int {
			a := 100
			f := func() *int {
				b := 200
				return &b
			}
			return 1
		}
	`, map[string]bool{
		"a": false,
		"b": true,
		"f": false,
	})
}
