package capture

import (
	"fmt"

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
	case *parser.MultiNameNode:
		for _, name := range n.Names {
			c.node(name)
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

	case *parser.CallNode:
		c.node(n.Function)
		c.nodes(n.Arguments)
	case *parser.OperatorNode:
		c.node(n.Left)
		c.node(n.Right)
	case *parser.ReturnNode:
		c.nodes(n.Vals)
	case *parser.AssignNode:
		c.nodes(n.Target)
		c.nodes(n.Val)
	case *parser.TypeCastNode:
		c.node(n.Val)
	case *parser.StructLoadElementNode:
		c.node(n.Struct)
	case *parser.LoadArrayElement:
		c.node(n.Array)
		c.node(n.Pos)
	case *parser.SliceArrayNode:
		c.node(n.Val)
		c.node(n.Start)
		c.node(n.End)
	case *parser.InitializeSliceNode:
		c.nodes(n.Items)
	case *parser.InitializeArrayNode:
		c.nodes(n.Items)
	case *parser.InitializeStructNode:
		for _, item := range n.Items {
			c.node(item)
		}

	case *parser.RangeNode:
		c.node(n.Item)
	case *parser.GetReferenceNode:
		c.node(n.Item)
	case *parser.DereferenceNode:
		c.node(n.Item)
	case *parser.NegateNode:
		c.node(n.Item)
	case *parser.SubNode:
		c.node(n.Item)
	case *parser.DeVariadicSliceNode:
		c.node(n.Item)
	case *parser.TypeCastInterfaceNode:
		c.node(n.Item)
	case *parser.DecrementNode:
		c.node(n.Item)
	case *parser.IncrementNode:
		c.node(n.Item)
	case *parser.GroupNode:
		c.node(n.Item)

	case *parser.ConstantNode, *parser.BreakNode, *parser.ContinueNode,
		*parser.DefineTypeNode, *parser.ImportNode, *parser.DeclarePackageNode:
		// nothing to do

	case parser.TypeNode:
		// types does not refer to variables

	default:
		panic(fmt.Sprintf("unexpected type in capture pass: %T", node))
	}
}
