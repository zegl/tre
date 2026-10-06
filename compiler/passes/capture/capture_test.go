package capture

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zegl/tre/compiler/lexer"
	"github.com/zegl/tre/compiler/parser"
)

type collector struct {
	allocs   map[string]bool
	literals []*parser.DefineFuncNode
	funcs    []*parser.DefineFuncNode
}

func (c *collector) Visit(node parser.Node) (parser.Node, parser.Visitor) {
	switch n := node.(type) {
	case *parser.AllocNode:
		for _, name := range n.Name {
			c.allocs[name] = n.Escapes
		}
	case *parser.DefineFuncNode:
		if n.IsNamed {
			c.funcs = append(c.funcs, n)
		} else {
			c.literals = append(c.literals, n)
		}
	}
	return node, c
}

func captureTest(t *testing.T, input string) *collector {
	parsed := parser.Parse(lexer.Lex(input), false)
	parsed = Capture(parsed)

	c := &collector{allocs: make(map[string]bool)}
	parser.Walk(c, parsed)
	return c
}

func TestNoCaptures(t *testing.T) {
	c := captureTest(t, `package main

	var global int

	func main() {
		a := 100
		f := func(b int) int {
			c := 1
			return b + c + global
		}
	}
`)
	assert.Equal(t, map[string]bool{"global": false, "a": false, "f": false, "c": false}, c.allocs)
	assert.Len(t, c.literals, 1)
	assert.Empty(t, c.literals[0].Captures)
}

func TestCaptureAlloc(t *testing.T) {
	c := captureTest(t, `package main

	func main() {
		a := 100
		b := 200
		f := func() int {
			a = a + 1
			return a
		}
	}
`)
	assert.Equal(t, map[string]bool{"a": true, "b": false, "f": false}, c.allocs)
	assert.Equal(t, []string{"a"}, c.literals[0].Captures)
}

func TestCaptureArgument(t *testing.T) {
	c := captureTest(t, `package main

	func counter(start int, step int) func() int {
		return func() int {
			start = start + step
			return start
		}
	}
`)
	assert.Equal(t, []string{"start", "step"}, c.literals[0].Captures)
	assert.Equal(t, map[string]bool{"start": true, "step": true}, c.funcs[0].EscapingArguments)
}

func TestCaptureNested(t *testing.T) {
	c := captureTest(t, `package main

	func main() {
		a := 100
		f := func() {
			b := 200
			g := func() int {
				return a + b
			}
		}
	}
`)
	assert.Equal(t, map[string]bool{"a": true, "b": true, "f": false, "g": false}, c.allocs)

	// f needs to capture a, to make it available to g
	assert.Equal(t, []string{"a"}, c.literals[0].Captures)
	assert.Equal(t, []string{"a", "b"}, c.literals[1].Captures)
}

func TestCaptureShadowed(t *testing.T) {
	c := captureTest(t, `package main

	func main() {
		a := 100
		f := func(a int) int {
			return a
		}
		g := func() int {
			a := 300
			return a
		}
	}
`)
	assert.False(t, c.allocs["a"])
	assert.Empty(t, c.literals[0].Captures)
	assert.Empty(t, c.literals[1].Captures)
}

func TestCaptureBlockScope(t *testing.T) {
	c := captureTest(t, `package main

	func main() {
		for i := 0; i < 3; i++ {
			x := i
			f := func() int {
				return i + x
			}
		}
	}
`)
	assert.True(t, c.allocs["i"])
	assert.True(t, c.allocs["x"])
	assert.Equal(t, []string{"i", "x"}, c.literals[0].Captures)
}

func TestCaptureRange(t *testing.T) {
	c := captureTest(t, `package main

	func main() {
		s := []int{1, 2, 3}
		for k, v := range s {
			f := func() int {
				return v
			}
		}
	}
`)
	assert.True(t, c.allocs["v"])
	assert.False(t, c.allocs["s"])
	assert.Equal(t, []string{"v"}, c.literals[0].Captures)
}
