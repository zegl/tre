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
					allocsChecked = append(allocsChecked, allocNode.Name)
					assert.Equal(t, expected[allocNode.Name], allocNode.Escapes, allocNode.Name)
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
