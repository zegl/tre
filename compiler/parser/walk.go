package parser

import "fmt"

type Visitor interface {
	Visit(node Node) (w Visitor)
}

func Walk(v Visitor, node Node) {
	if v = v.Visit(node); v == nil {
		return
	}

	switch n := node.(type) {
	case *FileNode:
		for _, a := range n.Instructions {
			Walk(v, a)
		}
	case *CallNode:
		for _, a := range n.Arguments {
			Walk(v, a)
		}
		Walk(v, n.Function)
	case *OperatorNode:
		Walk(v, n.Left)
		Walk(v, n.Right)
	case *ConstantNode:
		// nothing to do
	case *ConditionNode:
		Walk(v, n.Cond)
		for _, a := range n.True {
			Walk(v, a)
		}
		for _, a := range n.False {
			Walk(v, a)
		}
	case *DefineFuncNode:
		for _, a := range n.Body {
			Walk(v, a)
		}
	case *NameNode:
		// nothing to do
	case *MultiNameNode:
		// nothing to do
	case *ReturnNode:
		for _, a := range n.Vals {
			Walk(v, a)
		}
	case *AllocNode:
		for _, a := range n.Val {
			Walk(v, a)
		}
	case *AllocGroup:
		for _, a := range n.Allocs {
			Walk(v, a)
		}
	case *TypeCastNode:
		// nothing to do
	case *DefineTypeNode:
		// nothing to do
	case *StructLoadElementNode:
		// nothing to do
	case *LoadArrayElement:
		Walk(v, n.Array)
		Walk(v, n.Pos)
	case *SliceArrayNode:
		Walk(v, n.Start)
		Walk(v, n.End)
		Walk(v, n.Val)
	case *DeclarePackageNode:
		// nothing to do
	case *BreakNode:
		// nothing to do
	case *ContinueNode:
		// nothing to do
	case *GetReferenceNode:
		Walk(v, n.Item)
	case *DereferenceNode:
		Walk(v, n.Item)
	case *ImportNode:
		// nothing to do
	case *NegateNode:
		Walk(v, n.Item)
	case *SubNode:
		Walk(v, n.Item)
	case *InitializeSliceNode:
		for _, a := range n.Items {
			Walk(v, a)
		}
	case *InitializeArrayNode:
		for _, a := range n.Items {
			Walk(v, a)
		}
	case *DeVariadicSliceNode:
		Walk(v, n.Item)
	case *TypeCastInterfaceNode:
		Walk(v, n.Item)
	case *DecrementNode:
		Walk(v, n.Item)
	case *IncrementNode:
		Walk(v, n.Item)
	case *GroupNode:
		Walk(v, n.Item)
	case *ForNode:
		Walk(v, n.BeforeLoop)
		Walk(v, n.Condition)
		Walk(v, n.AfterIteration)
		for _, a := range n.Block {
			Walk(v, a)
		}
	case *RangeNode:
		Walk(v, n.Item)
	case *SwitchNode:
		Walk(v, n.Item)
		for _, a := range n.Cases {
			Walk(v, a)
		}
		for _, a := range n.DefaultBody {
			Walk(v, a)
		}
	case *SwitchCaseNode:
		for _, a := range n.Conditions {
			Walk(v, a)
		}
		for _, a := range n.Body {
			Walk(v, a)
		}
	case *AssignNode:
		for _, a := range n.Target {
			Walk(v, a)
		}
		for _, a := range n.Val {
			Walk(v, a)
		}
	default:
		panic(fmt.Sprintf("unexpected type in Walk(): %T", node))
	}
}
