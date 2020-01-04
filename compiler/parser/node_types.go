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
	SetName(string)
	GetName() string
}

// SingleTypeNode refers to an existing type. Such as "string".
type SingleTypeNode struct {
	baseNode

	PackageName string
	SourceName  string
	TypeName    string
	IsVariadic  bool
}

func (stn SingleTypeNode) Type() string {
	return stn.TypeName
}

func (stn SingleTypeNode) String() string {
	return fmt.Sprintf("type(%s.%s)", stn.PackageName, stn.Type())
}

func (stn SingleTypeNode) Variadic() bool {
	return stn.IsVariadic
}

func (stn SingleTypeNode) SetName(name string) {
	stn.SourceName = name
}

func (stn SingleTypeNode) GetName() string {
	return stn.SourceName
}

// StructTypeNode refers to a struct type
type StructTypeNode struct {
	baseNode

	SourceName string
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

func (stn StructTypeNode) SetName(name string) {
	stn.SourceName = name
}

func (stn StructTypeNode) GetName() string {
	return stn.SourceName
}

// ArrayTypeNode refers to an array
type ArrayTypeNode struct {
	baseNode

	SourceName string
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

func (atn ArrayTypeNode) SetName(name string) {
	atn.SourceName = name
}

func (atn ArrayTypeNode) GetName() string {
	return atn.SourceName
}

type SliceTypeNode struct {
	baseNode

	SourceName string
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

func (stn SliceTypeNode) SetName(name string) {
	stn.SourceName = name
}

func (stn SliceTypeNode) GetName() string {
	return stn.SourceName
}

type InterfaceTypeNode struct {
	baseNode

	SourceName string
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

func (itn InterfaceTypeNode) SetName(name string) {
	itn.SourceName = name
}

func (itn InterfaceTypeNode) GetName() string {
	return itn.SourceName
}

type InterfaceMethod struct {
	ArgumentTypes []TypeNode
	ReturnTypes   []TypeNode
}

type PointerTypeNode struct {
	baseNode

	SourceName string
	IsVariadic bool
	ValueType  TypeNode
}

func (ptn PointerTypeNode) Type() string {
	return fmt.Sprintf("pointer(%+v)", ptn.ValueType.Type())
}

func (ptn PointerTypeNode) String() string {
	return ptn.Type()
}

func (ptn PointerTypeNode) SetName(name string) {
	ptn.SourceName = name
}

func (ptn PointerTypeNode) GetName() string {
	return ptn.SourceName
}

func (ptn PointerTypeNode) Variadic() bool {
	return ptn.IsVariadic
}

type FuncTypeNode struct {
	baseNode

	ArgTypes []TypeNode
	RetTypes []TypeNode

	SourceName string
	IsVariadic bool
}

func (ftn FuncTypeNode) Type() string {
	return fmt.Sprintf("func(%+v)(%+v)", ftn.ArgTypes, ftn.RetTypes)
}

func (ftn FuncTypeNode) String() string {
	return ftn.Type()
}

func (ftn FuncTypeNode) SetName(name string) {
	ftn.SourceName = name
}

func (ftn FuncTypeNode) GetName() string {
	return ftn.SourceName
}

func (ftn FuncTypeNode) Variadic() bool {
	return ftn.IsVariadic
}
