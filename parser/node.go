package parser

import "fmt"

type Node interface {
}

// Function calls
type CallNode struct {
	Function  string
	Arguments []Node
}

func (cn CallNode) String() string {
	return fmt.Sprintf("CallNode: %s(%+v)", cn.Function, cn.Arguments)
}

// A block, consists only of other instructions
type BlockNode struct {
	Instructions []Node
}

func (bn BlockNode) String() string {
	return fmt.Sprintf("BlockNode: %+v", bn.Instructions)
}

// Math operations
type OperatorNode struct {
	Operator Operator
	Left     Node
	Right    Node
}

type Operator string

const (
	OP_ADD Operator = "+"
	OP_SUB          = "-"
	OP_DIV          = "/"
	OP_MUL          = "*"

	OP_GT   = ">"
	OP_GTEQ = ">="
	OP_LT   = "<"
	OP_LTEQ = "<="
	OP_EQ   = "=="
	OP_NEQ  = "!="
)

var opsCharToOp map[string]Operator

func init() {
	opsCharToOp = make(map[string]Operator)
	for _, op := range []Operator{
		OP_ADD, OP_SUB, OP_DIV, OP_MUL,
		OP_GT, OP_GTEQ, OP_LT, OP_LTEQ,
		OP_EQ, OP_NEQ,
	} {
		opsCharToOp[string(op)] = op
	}
}

func (on OperatorNode) String() string {
	return fmt.Sprintf("(%v %s %v)", on.Left, string(on.Operator), on.Right)
}

// Constants, strings and numbers
type ConstantNode struct {
	Type     DataType
	Value    int64
	ValueStr string
}

type DataType uint8

const (
	STRING DataType = iota
	NUMBER
)

func (cn ConstantNode) String() string {
	return fmt.Sprintf("%d", cn.Value)
}

// If conditions
type ConditionNode struct {
	Cond  OperatorNode
	True  []Node
	False []Node
}

type DefineFuncNode struct {
	Name         string
	Arguments    []NameNode
	ReturnValues []NameNode
	Body         []Node
}

// Variables, etc.
type NameNode struct {
	Name string
	Type string
}

type ReturnNode struct {
	Val Node
}

func (rn ReturnNode) String() string {
	return fmt.Sprintf("return %v", rn.Val)
}

type AllocNode struct {
	Name string
	Val  Node
}

func (an AllocNode) String() string {
	return fmt.Sprintf("alloc %s = %v", an.Name, an.Val)
}

type AssignNode struct {
	Name string
	Val  Node
}

func (an AssignNode) String() string {
	return fmt.Sprintf("assign %s = %v", an.Name, an.Val)
}

type TypeCastNode struct {
	Type string
	Val  Node
}

func (tcn TypeCastNode) String() string {
	return fmt.Sprintf("cast %s(%v)", tcn.Type, tcn.Val)
}

type DefineTypeNode struct {
	Name string
	Type TypeNode
}

type TypeNode interface {
	Type() string
}

type SingleTypeNode struct {
	TypeName string
}

func (stn *SingleTypeNode) Type() string {
	return stn.TypeName
}

type StructTypeNode struct {
	Types []TypeNode
	Names map[string]int
}

func (stn *StructTypeNode) Type() string {
	return fmt.Sprintf("%+v", stn.Types)
}
