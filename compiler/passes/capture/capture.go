package capture

import (
	"github.com/zegl/tre/compiler/parser"
)

// Capture finds the variables that anonymous functions use from their
// enclosing functions.
//
// The variables are added to Captures of every function literal between the
// use and the declaration of the variable. The declaration is marked as
// escaping, as the variable can outlive the function that declared it.
func Capture(input *parser.FileNode) *parser.FileNode {
	for _, ins := range input.Instructions {
		if defFunc, ok := ins.(*parser.DefineFuncNode); ok {
			c := &capturer{}
			c.function(defFunc)
		}
	}

	return input
}

// declaration is where a variable was declared
type declaration struct {
	// Set when declared by an AllocNode
	alloc *parser.AllocNode

	// Set when declared as an argument to a function
	argumentOf *parser.DefineFuncNode
}

func (d declaration) markEscaping(name string) {
	if d.alloc != nil {
		d.alloc.Escapes = true
	}

	if d.argumentOf != nil {
		if d.argumentOf.EscapingArguments == nil {
			d.argumentOf.EscapingArguments = make(map[string]bool)
		}
		d.argumentOf.EscapingArguments[name] = true
	}
}

type function struct {
	node *parser.DefineFuncNode

	// Stack of block scopes
	scopes []map[string]declaration
}

type capturer struct {
	// Stack of functions, the innermost function is last
	funcs []*function
}

func (c *capturer) current() *function {
	return c.funcs[len(c.funcs)-1]
}

func (c *capturer) pushScope() {
	f := c.current()
	f.scopes = append(f.scopes, make(map[string]declaration))
}

func (c *capturer) popScope() {
	f := c.current()
	f.scopes = f.scopes[:len(f.scopes)-1]
}

func (c *capturer) declare(name string, decl declaration) {
	f := c.current()
	f.scopes[len(f.scopes)-1][name] = decl
}

// use resolves a variable name, and records a capture if the variable was
// declared in an enclosing function
func (c *capturer) use(name string) {
	for fi := len(c.funcs) - 1; fi >= 0; fi-- {
		f := c.funcs[fi]

		for si := len(f.scopes) - 1; si >= 0; si-- {
			decl, ok := f.scopes[si][name]
			if !ok {
				continue
			}

			// Declared in the current function
			if fi == len(c.funcs)-1 {
				return
			}

			decl.markEscaping(name)

			// All functions between the declaration and the use needs to
			// capture the variable
			for _, inner := range c.funcs[fi+1:] {
				addCapture(inner.node, name)
			}
			return
		}
	}

	// Not a local variable, the name refers to something on the package level
}

func addCapture(fn *parser.DefineFuncNode, name string) {
	for _, existing := range fn.Captures {
		if existing == name {
			return
		}
	}
	fn.Captures = append(fn.Captures, name)
}

func (c *capturer) function(fn *parser.DefineFuncNode) {
	c.funcs = append(c.funcs, &function{node: fn})
	c.pushScope()

	if fn.IsMethod {
		c.declare(fn.InstanceName, declaration{argumentOf: fn})
	}
	for _, arg := range fn.Arguments {
		c.declare(arg.Name, declaration{argumentOf: fn})
	}

	// Named return values are not moved to the heap
	for _, ret := range fn.ReturnValues {
		if ret.Name != "" {
			c.declare(ret.Name, declaration{})
		}
	}

	c.block(fn.Body)

	c.popScope()
	c.funcs = c.funcs[:len(c.funcs)-1]
}

func (c *capturer) block(nodes []parser.Node) {
	c.pushScope()
	c.nodes(nodes)
	c.popScope()
}

func (c *capturer) nodes(nodes []parser.Node) {
	for _, n := range nodes {
		c.node(n)
	}
}

func (c *capturer) alloc(n *parser.AllocNode) {
	// The values are evaluated before the new variables are in scope
	c.nodes(n.Val)

	for _, name := range n.Name {
		c.declare(name, declaration{alloc: n})
	}
}

func (c *capturer) node(node parser.Node) {
	switch n := node.(type) {
	case nil:
		// nothing to do

	case *parser.DefineFuncNode:
		c.function(n)

	case *parser.AllocNode:
		c.alloc(n)
	case *parser.AllocGroup:
		for _, a := range n.Allocs {
			c.alloc(a)
		}

	case *parser.NameNode:
		if n.Package == "" {
			c.use(n.Name)
		}
	case *parser.ConditionNode:
		c.node(n.Cond)
		c.block(n.True)
		c.block(n.False)
	case *parser.ForNode:
		c.pushScope()
		c.node(n.BeforeLoop)
		if n.Condition != nil {
			c.node(n.Condition)
		}
		c.node(n.AfterIteration)
		c.block(n.Block)
		c.popScope()
	case *parser.SwitchNode:
		c.pushScope()
		c.node(n.Item)
		for _, switchCase := range n.Cases {
			c.nodes(switchCase.Conditions)
			c.block(switchCase.Body)
		}
		c.block(n.DefaultBody)
		c.popScope()

	default:
		// Expressions and statements that does not declare any variables
		c.nodes(parser.Children(node))
	}
}
