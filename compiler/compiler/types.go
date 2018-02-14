package compiler

import (
	"fmt"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"

	"github.com/zegl/tre/compiler/parser"
)

var typeMethods = map[string]map[string]*ir.Function{}

var typeConvertMap = map[string]types.Type{
	"int":   types.I64, // TODO: Size based on arch
	"int8":  types.I8,
	"int32": types.I32,
	"int64": types.I64,
}

// Type Name : Element Name : Index
var typeMapElementNameIndex = map[string]map[string]int{}

// Type Name : Method Name : Function
var typeMapMethodNameFunction = map[string]map[string]method{}

type method struct {
	Func *ir.Function
	// TODO: Implement case where this is true
	PointerReceiver bool
}

// methodCall represents a method call
// method is the method to be called
// Value is a instance of the type that has the method
type methodCall struct {
	value.Value
	method method
}

func typeStringToLLVM(sourceName string) types.Type {
	if t, ok := typeConvertMap[sourceName]; ok {
		return t
	}

	panic("unknown type: " + sourceName)
}

func typeNodeToLLVMType(typeNode parser.TypeNode) types.Type {
	switch t := typeNode.(type) {
	case parser.SingleTypeNode:
		return typeStringToLLVM(t.TypeName)

	case parser.ArrayTypeNode:
		return types.NewArray(typeNodeToLLVMType(t.ItemType), t.Len)

	case parser.StructTypeNode:
		var structTypes []types.Type
		for _, tt := range t.Types {
			structTypes = append(structTypes, typeNodeToLLVMType(tt))
		}
		return types.NewStruct(structTypes...)
	}

	panic(fmt.Sprintf("unknown typeNode: %T", typeNode))
}
