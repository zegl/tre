package parser

import (
	"fmt"
	"strings"
)

// Node is the base node. A node consists of something that the compiler or language can do
type Node interface {
	Node()
	String() string
}

// baseNode implements the Node interface, to recuce code duplication
type baseNode struct{}

func (n *baseNode) Node() {

}

// CallNode is a function call. Function is the name of the function to execute.
type CallNode struct {
	baseNode

	Function  Node
	Arguments []Node
}

func (cn CallNode) String() string {
	return fmt.Sprintf("CallNode: %s(%+v)", cn.Function, cn.Arguments)
}

type PackageNode struct {
	baseNode

	Files []FileNode
	Name  string
}

// FileNode is a list of other nodes
// Indicates the root of one source file
type FileNode struct {
	baseNode

	Instructions []Node
}

func (fn FileNode) String() string {
	var res []string

	for _, i := range fn.Instructions {
		res = append(res, fmt.Sprintf("%+v", i))
	}

	return fmt.Sprintf("FileNode: \n\t%s", strings.Join(res, "\n\t"))
}

// OperatorNode is mathematical operations and comparisons
type OperatorNode struct {
	baseNode

	Operator Operator
	Left     Node
	Right    Node
}

type Operator string

// https://golang.org/ref/spec#Arithmetic_operators
const (
	OP_ADD       Operator = "+"
	OP_SUB                = "-"
	OP_DIV                = "/"
	OP_MUL                = "*"
	OP_REMAINDER          = "%"

	OP_BIT_AND   = "&"
	OP_BIT_OR    = "|"
	OP_BIT_XOR   = "^"
	OP_BIT_CLEAR = "&^" // AND NOT

	OP_LEFT_SHIFT  = "<<"
	OP_RIGHT_SHIFT = ">>"

	OP_GT   = ">"
	OP_GTEQ = ">="
	OP_LT   = "<"
	OP_LTEQ = "<="
	OP_EQ   = "=="
	OP_NEQ  = "!="

	OP_LOGICAL_AND = "&&"
	OP_LOGICAL_OR  = "||"
)

var opsCharToOp map[string]Operator
var arithOperators map[Operator]struct{}

func init() {
	opsCharToOp = make(map[string]Operator)
	for _, op := range []Operator{
		OP_ADD, OP_SUB, OP_DIV, OP_MUL, OP_REMAINDER,
		OP_BIT_AND, OP_BIT_OR, OP_BIT_XOR, OP_BIT_CLEAR,
		OP_LEFT_SHIFT, OP_RIGHT_SHIFT,
		OP_GT, OP_GTEQ, OP_LT, OP_LTEQ, OP_EQ, OP_NEQ,
	} {
		opsCharToOp[string(op)] = op
	}

	arithOperators = make(map[Operator]struct{})
	for _, op := range []Operator{
		OP_ADD, OP_SUB, OP_DIV, OP_MUL,
	} {
		arithOperators[op] = struct{}{}
	}
}

func (on OperatorNode) String() string {
	return fmt.Sprintf("OP(%v %s %v)", on.Left, string(on.Operator), on.Right)
}

// ConstantNode is a raw string or number
type ConstantNode struct {
	baseNode

	Type     DataType
	Value    int64
	ValueStr string
}

type DataType uint8

const (
	STRING DataType = iota
	NUMBER
	BOOL
)

func (cn ConstantNode) String() string {
	if len(cn.ValueStr) > 0 {
		return cn.ValueStr
	}
	return fmt.Sprintf("const(%d)", cn.Value)
}

// ConditionNode creates a new if condition
// False is optional
type ConditionNode struct {
	baseNode

	Cond  *OperatorNode
	True  []Node
	False []Node
}

func (ConditionNode) String() string {
	return "condition"
}

// DefineFuncNode creates a new named function
type DefineFuncNode struct {
	baseNode

	Name    string
	IsNamed bool

	IsMethod bool

	MethodOnType      *SingleTypeNode
	IsPointerReceiver bool
	InstanceName      string

	Arguments    []*NameNode
	ReturnValues []*NameNode
	Body         []Node
}

func (dfn DefineFuncNode) String() string {
	var body []string

	for _, b := range dfn.Body {
		body = append(body, fmt.Sprintf("%+v", b))
	}

	if dfn.IsMethod {
		return fmt.Sprintf("func+m (%+v) %s(%+v) %+v {\n\t%s\n}", dfn.InstanceName, dfn.Name, dfn.Arguments, dfn.ReturnValues, strings.Join(body, "\n\t"))
	} else if dfn.IsNamed {
		return fmt.Sprintf("func+n %s(%+v) %+v {\n\t%s\n}", dfn.Name, dfn.Arguments, dfn.ReturnValues, strings.Join(body, "\n\t"))
	} else {
		return fmt.Sprintf("func+v (%+v) %+v {\n\t%s\n}", dfn.Arguments, dfn.ReturnValues, strings.Join(body, "\n\t"))
	}
}

// NameNode retreives a named variable
type NameNode struct {
	baseNode

	Package string
	Name    string
	Type    TypeNode
}

func (nn NameNode) String() string {
	if nn.Type == nil {
		return fmt.Sprintf("var(%s·%s)", nn.Package, nn.Name)
	}
	return fmt.Sprintf("var(n:%s·%s t:%s)", nn.Package, nn.Name, nn.Type)
}

type MultiNameNode struct {
	baseNode
	Names []*NameNode
}

func (n MultiNameNode) String() string {
	return fmt.Sprintf("vars(%+v)", n.Names)
}

// ReturnNode returns Val from within the current function
type ReturnNode struct {
	baseNode

	Vals []Node
}

func (rn ReturnNode) String() string {
	return fmt.Sprintf("return(%v)", rn.Vals)
}

// AllocNode creates a new variable Name with the value Val
type AllocNode struct {
	baseNode

	Escapes bool

	Name []string
	Val  []Node

	Type TypeNode // Is set when allocating on the format "var ident int" and "var ident, ident int = expr, expr"

	IsConst bool // Is true when in a const expression.
}

func (an AllocNode) String() string {
	return fmt.Sprintf("alloc(%s) = %v (escapes: %v type: %v)", an.Name, an.Val, an.Escapes, an.Type)
}

type AllocGroup struct {
	baseNode
	Allocs []*AllocNode
}

func (an AllocGroup) String() string {
	return fmt.Sprintf("alloc(%+v)", an.Allocs)
}

// AssignNode assign Val to Target (or Name)
type AssignNode struct {
	baseNode

	Target []Node
	Val    []Node
}

func (an AssignNode) String() string {
	return fmt.Sprintf("assign(%+v) = %v", an.Target, an.Val)
}

// TypeCastNode tries to cast Val to Type
type TypeCastNode struct {
	baseNode

	Type TypeNode
	Val  Node
}

func (tcn TypeCastNode) String() string {
	return fmt.Sprintf("cast %s(%v)", tcn.Type, tcn.Val)
}

// DefineTypeNode creates a new named type
type DefineTypeNode struct {
	baseNode

	Name string
	Type TypeNode
}

func (dtn DefineTypeNode) String() string {
	return fmt.Sprintf("defineType %s = %+v", dtn.Name, dtn.Type)
}

// StructLoadElementNode retrieves a value by key from a struct
type StructLoadElementNode struct {
	baseNode

	Struct      Node
	ElementName string
}

func (slen StructLoadElementNode) String() string {
	return fmt.Sprintf("load(%+v.%+v)", slen.Struct, slen.ElementName)
}

// LoadArrayElement loads a single element from an array
// On the form arr[1]
type LoadArrayElement struct {
	baseNode

	Array Node
	Pos   Node
}

func (l LoadArrayElement) String() string {
	return fmt.Sprintf("loadArrayElement(%+v[%+v])", l.Array, l.Pos)
}

// SliceArrayNode slices an array or string
// Can be on the forms arr[1], arr[1:], or arr[1:3]
type SliceArrayNode struct {
	baseNode

	Val    Node
	Start  Node
	HasEnd bool
	End    Node
}

func (san SliceArrayNode) String() string {
	return fmt.Sprintf("slice(%+v[%s:%s])", san.Val.String(), san.Start.String(), san.End.String())
}

// DeclarePackageNode declares the package that we're in
type DeclarePackageNode struct {
	baseNode

	PackageName string
}

func (d DeclarePackageNode) String() string {
	return "DeclarePackageNode(" + d.PackageName + ")"
}

// BreakNode breaks out of the current for loop
type BreakNode struct {
	baseNode
}

func (n BreakNode) String() string {
	return "break"
}

// ContinueNode skips the current iteration of the current for loop
type ContinueNode struct {
	baseNode
}

func (n ContinueNode) String() string {
	return "continue"
}

type GetReferenceNode struct {
	baseNode
	Item Node
}

func (grn GetReferenceNode) String() string {
	return fmt.Sprintf("&(%s)", grn.Item)
}

type DereferenceNode struct {
	baseNode
	Item Node
}

func (dn DereferenceNode) String() string {
	return fmt.Sprintf("*(%s)", dn.Item)
}

type ImportNode struct {
	baseNode
	PackagePaths []string
}

func (in ImportNode) String() string {
	return fmt.Sprintf("import (%s)", strings.Join(in.PackagePaths, ", "))
}

type NegateNode struct {
	baseNode
	Item Node
}

func (nn NegateNode) String() string {
	return fmt.Sprintf("NegateNode-!%s", nn.Item)
}

type SubNode struct {
	baseNode
	Item Node
}

func (sn SubNode) String() string {
	return fmt.Sprintf("-%s", sn.Item)
}

type InitializeSliceNode struct {
	baseNode
	Type  TypeNode
	Items []Node
}

func (i InitializeSliceNode) String() string {
	return fmt.Sprintf("InitializeSliceNode-[]%s{%+v}", i.Type, i.Items)
}

type InitializeArrayNode struct {
	baseNode
	Type  TypeNode
	Size  int
	Items []Node
}

func (i InitializeArrayNode) String() string {
	return fmt.Sprintf("InitializeArrayNode-[%d]%s{%+v}", i.Size, i.Type, i.Items)
}

type InitializeStructNode struct {
	baseNode
	Type  TypeNode
	Items map[string]Node
}

func (i InitializeStructNode) String() string {
	return fmt.Sprintf("InitializeStructNode-%s{%+v}", i.Type, i.Items)
}

type DeVariadicSliceNode struct {
	baseNode
	Item Node
}

func (i DeVariadicSliceNode) String() string {
	return fmt.Sprintf("%+v...", i.Item)
}

type TypeCastInterfaceNode struct {
	baseNode
	Item Node
	Type TypeNode
}

func (i TypeCastInterfaceNode) String() string {
	return fmt.Sprintf("castInterface(%s(%+v))", i.Type, i.Item)
}

type DecrementNode struct {
	baseNode
	Item Node
}

func (i DecrementNode) String() string {
	return fmt.Sprintf("%s--", i.Item)
}

type IncrementNode struct {
	baseNode
	Item Node
}

func (i IncrementNode) String() string {
	return fmt.Sprintf("%s++", i.Item)
}

type GroupNode struct {
	baseNode
	Item Node
}

func (i GroupNode) String() string {
	return fmt.Sprintf("( %s )", i.Item)
}
