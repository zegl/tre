package escape

import (
	"github.com/zegl/tre/compiler/parser"
)

// Escape performs variable escape analysis on variables allocated in functions
func Escape(input *parser.FileNode) *parser.FileNode {
	for _, ins := range input.Instructions {
		if defFunc, ok := ins.(*parser.DefineFuncNode); ok {
			escapeFunc(defFunc)
		}
	}

	return input
}

func escapeFunc(defFunc *parser.DefineFuncNode) {
	// Name of the var mapped to the allocNodes that declares it
	allocatedVars := map[string][]*parser.AllocNode{}
	escapingVars := map[string]struct{}{}

	var visit func(node parser.Node)
	visit = func(node parser.Node) {
		switch n := node.(type) {
		case *parser.DefineFuncNode:
			// Function literals are analyzed as separate functions
			escapeFunc(n)
			return

		case *parser.AllocNode:
			// Find all variables allocated in this function
			for _, name := range n.Name {
				allocatedVars[name] = append(allocatedVars[name], n)
			}

		case *parser.ReturnNode:
			// Find all variables returned from this function
			for _, val := range n.Vals {
				findEscaping(escapingVars, val)
			}
		}

		for _, child := range parser.Children(node) {
			visit(child)
		}
	}

	for _, ins := range defFunc.Body {
		visit(ins)
	}

	// Mark as escaping in the AST
	for escapingName := range escapingVars {
		for _, allocIns := range allocatedVars[escapingName] {
			allocIns.Escapes = true
		}
	}
}

func findEscaping(escapingVars map[string]struct{}, ins parser.Node) {
	if retVariable, ok := ins.(*parser.NameNode); ok {
		escapingVars[retVariable.Name] = struct{}{}
		return
	}

	if retPtr, ok := ins.(*parser.GetReferenceNode); ok {
		findEscaping(escapingVars, retPtr.Item)
		return
	}

	/*if initStruct, ok := ins.(*parser.InitializeStructNode); ok {
		// initStruct.Items
	}*/
}
