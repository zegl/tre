package parser

import (
	"fmt"
)

// TypeNode is an interface for different ways of creating new types or referring to existing ones
type TypeNode interface {
	Node() // must also implement the Node interface
	Type() string
	String() string
	Variadic() bool
}

// SingleTypeNode refers to an existing type. Such as "string".
type SingleTypeNode struct {
	baseNode

	TypeName   string
	IsVariadic bool
}

func (stn SingleTypeNode) Type() string {
	return stn.TypeName
}

func (stn SingleTypeNode) String() string {
	return "type(" + stn.Type() + ")"
}

func (stn SingleTypeNode) Variadic() bool {
	return stn.IsVariadic
}

// StructTypeNode refers to a struct type
type StructTypeNode struct {
	baseNode

	Types      []TypeNode
	Names      map[string]int
	IsVariadic bool
}

func (stn StructTypeNode) Type() string {
	return fmt.Sprintf("%+v", stn.Types)
}

func (stn StructTypeNode) String() string {
	return fmt.Sprintf("StructTypeNode(%+v)", stn.Types)
}

func (stn StructTypeNode) Variadic() bool {
	return stn.IsVariadic
}

// ArrayTypeNode refers to an array
type ArrayTypeNode struct {
	baseNode

	ItemType   TypeNode
	Len        int64
	IsVariadic bool
}

func (atn ArrayTypeNode) Type() string {
	return fmt.Sprintf("[%d]%+v", atn.Len, atn.ItemType)
}

func (atn ArrayTypeNode) String() string {
	return atn.Type()
}

func (atn ArrayTypeNode) Variadic() bool {
	return atn.IsVariadic
}

type SliceTypeNode struct {
	baseNode
	ItemType   TypeNode
	IsVariadic bool
}

func (stn SliceTypeNode) Type() string {
	return fmt.Sprintf("[]%+v", stn.ItemType)
}

func (stn SliceTypeNode) String() string {
	return stn.Type()
}

func (stn SliceTypeNode) Variadic() bool {
	return stn.IsVariadic
}

type InterfaceTypeNode struct {
	baseNode

	Methods    map[string]InterfaceMethod
	IsVariadic bool
}

func (itn InterfaceTypeNode) Type() string {
	return fmt.Sprintf("interface{%+v}", itn.Methods)
}

func (itn InterfaceTypeNode) String() string {
	return itn.Type()
}

func (itn InterfaceTypeNode) Variadic() bool {
	return itn.IsVariadic
}

type InterfaceMethod struct {
	ArgumentTypes []TypeNode
	ReturnTypes   []TypeNode
}
