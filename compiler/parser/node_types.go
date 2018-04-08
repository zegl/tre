package parser

import "fmt"

// TypeNode is an interface for different ways of creating new types or referring to existing ones
type TypeNode interface {
	Node() // must also implement the Node interface
	Type() string
	String() string
}

// SingleTypeNode refers to an existing type. Such as "string".
type SingleTypeNode struct {
	baseNode

	TypeName string
}

func (stn SingleTypeNode) Type() string {
	return stn.TypeName
}

func (stn SingleTypeNode) String() string {
	return "type(" + stn.Type() + ")"
}

// StructTypeNode refers to a struct type
type StructTypeNode struct {
	baseNode

	Types []TypeNode
	Names map[string]int
}

func (stn StructTypeNode) Type() string {
	return fmt.Sprintf("%+v", stn.Types)
}

func (stn StructTypeNode) String() string {
	return fmt.Sprintf("StructTypeNode(%+v)", stn.Types)
}

// ArrayTypeNode refers to an array
type ArrayTypeNode struct {
	baseNode

	ItemType TypeNode
	Len      int64
}

func (atn ArrayTypeNode) Type() string {
	return fmt.Sprintf("[%d]%+v", atn.Len, atn.ItemType)
}

func (atn ArrayTypeNode) String() string {
	return atn.Type()
}

type SliceTypeNode struct {
	baseNode
	ItemType TypeNode
}

func (stn SliceTypeNode) Type() string {
	return fmt.Sprintf("[]%+v", stn.ItemType)
}

func (stn SliceTypeNode) String() string {
	return stn.Type()
}
