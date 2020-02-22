package parser

import (
	"fmt"
)

type Visitor interface {
	Visit(node Node) (n Node, w Visitor)
}

func Walk(v Visitor, node Node) (r Node) {
	if node, v = v.Visit(node); v == nil {
		return node
	}
	r = node

	if node == nil {
		return
	}

	switch n := node.(type) {
	case *FileNode:
		for i, a := range n.Instructions {
			n.Instructions[i] = Walk(v, a)
		}
	case *CallNode:
		for i, a := range n.Arguments {
			n.Arguments[i] = Walk(v, a)
		}
		n.Function = Walk(v, n.Function)
	case *OperatorNode:
		n.Left = Walk(v, n.Left)
		n.Right = Walk(v, n.Right)
	case *ConstantNode:
		// nothing to do
	case *ConditionNode:
		n.Cond = Walk(v, n.Cond).(*OperatorNode)
		for i, a := range n.True {
			n.True[i] = Walk(v, a)
		}
		for i, a := range n.False {
			n.False[i] = Walk(v, a)
		}
	case *DefineFuncNode:
		for i, a := range n.Body {
			n.Body[i] = Walk(v, a)
		}
	case *NameNode:
		// nothing to do
	case *MultiNameNode:
		// nothing to do
	case *ReturnNode:
		for i, a := range n.Vals {
			n.Vals[i] = Walk(v, a)
		}
	case *AllocNode:
		for i, a := range n.Val {
			n.Val[i] = Walk(v, a)
		}
	case *AllocGroup:
		for i, a := range n.Allocs {
			n.Allocs[i] = Walk(v, a).(*AllocNode)
		}
	case *TypeCastNode:
		// nothing to do
	case *DefineTypeNode:
		// nothing to do
	case *StructLoadElementNode:
		// nothing to do
	case *LoadArrayElement:
		n.Array = Walk(v, n.Array)
		n.Pos = Walk(v, n.Pos)
	case *SliceArrayNode:
		n.Start = Walk(v, n.Start)
		n.End = Walk(v, n.End)
		n.Val = Walk(v, n.Val)
	case *DeclarePackageNode:
		// nothing to do
	case *BreakNode:
		// nothing to do
	case *ContinueNode:
		// nothing to do
	case *GetReferenceNode:
		n.Item = Walk(v, n.Item)
	case *DereferenceNode:
		n.Item = Walk(v, n.Item)
	case *ImportNode:
		// nothing to do
	case *NegateNode:
		n.Item = Walk(v, n.Item)
	case *SubNode:
		n.Item = Walk(v, n.Item)
	case *InitializeSliceNode:
		for i, a := range n.Items {
			n.Items[i] = Walk(v, a)
		}
	case *InitializeArrayNode:
		for i, a := range n.Items {
			n.Items[i] = Walk(v, a)
		}
	case *DeVariadicSliceNode:
		n.Item = Walk(v, n.Item)
	case *TypeCastInterfaceNode:
		n.Item = Walk(v, n.Item)
	case *DecrementNode:
		n.Item = Walk(v, n.Item)
	case *IncrementNode:
		n.Item = Walk(v, n.Item)
	case *GroupNode:
		n.Item = Walk(v, n.Item)
	case *ForNode:
		n.BeforeLoop = Walk(v, n.BeforeLoop)
		if n.Condition != nil {
			if cond, ok := Walk(v, n.Condition).(*OperatorNode); ok {
				n.Condition = cond
			}
		}
		n.AfterIteration = Walk(v, n.AfterIteration)
		for i, a := range n.Block {
			n.Block[i] = Walk(v, a)
		}
	case *RangeNode:
		n.Item = Walk(v, n.Item)
	case *SwitchNode:
		n.Item = Walk(v, n.Item)
		for i, a := range n.Cases {
			n.Cases[i] = Walk(v, a).(*SwitchCaseNode)
		}
		for i, a := range n.DefaultBody {
			n.DefaultBody[i] = Walk(v, a)
		}
	case *SwitchCaseNode:
		for i, a := range n.Conditions {
			n.Conditions[i] = Walk(v, a)
		}
		for i, a := range n.Body {
			n.Body[i] = Walk(v, a)
		}
	case *AssignNode:
		for i, a := range n.Target {
			n.Target[i] = Walk(v, a)
		}
		for i, a := range n.Val {
			n.Val[i] = Walk(v, a)
		}
	case *InitializeStructNode:
		for i, a := range n.Items {
			n.Items[i] = Walk(v, a)
		}
	default:
		panic(fmt.Sprintf("unexpected type in Walk(): %T", node))
	}
	return
}
