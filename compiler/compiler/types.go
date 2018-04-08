package compiler

import (
	"fmt"

	"github.com/zegl/tre/compiler/compiler/internal"

	"github.com/zegl/tre/compiler/compiler/types"
	"github.com/zegl/tre/compiler/parser"

	llvmTypes "github.com/llir/llvm/ir/types"
)

var typeConvertMap = map[string]types.Type{
	"int":    types.I64, // TODO: Size based on arch
	"int8":   types.I8,
	"int16":  types.I16,
	"int32":  types.I32,
	"int64":  types.I64,
	"string": types.String,
}

// Type Name : Element Name : Index
// var typeMapElementNameIndex = map[string]map[string]int{}
/*var typeStructMaps = map[string]treStruct{}

// Type Name : Method Name : Function
var typeMapMethodNameFunction = map[string]map[string]method{}

var structMap = map[string]treStruct{}

type treStruct struct {
	elements map[string]structElement
	methods  map[string]method
}

type structElement struct {
	datatype datatype
	index    int
}

type function struct {
	*ir.Function
	retType datatype
}

type method struct {
	Func function
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

type variable struct {
	value.Value
	datatype datatype
}

type datatype struct {
	types.Type

	baseType string

	// keeping track of how many levels deep in pointers we are
	// *int = 1
	// **int = 2
	pointerLevel uint

	// TODO: It should be possible to design this in a much nicer way
	isStruct   bool
	structType treStruct
}

func valueToVariable(val value.Value, pointerLevel uint) variable {
	return variable{
		Value: val,
		datatype: datatype{
			Type:         val.Type(),
			baseType:     val.Type().String(),
			pointerLevel: pointerLevel,
		},
	}
}

func typeStringToLLVM(sourceName string) types.Type {
	if t, ok := typeConvertMap[sourceName]; ok {
		return t
	}

	panic("unknown type: " + sourceName)
}
*/

func parserTypeToType(typeNode parser.TypeNode) types.Type {
	switch t := typeNode.(type) {
	case parser.SingleTypeNode:
		// return datatype{
		// 	Type:         typeStringToLLVM(t.TypeName),
		// 	baseType:     t.TypeName,
		// 	pointerLevel: 0,
		// }
		if res, ok := typeConvertMap[t.TypeName]; ok {
			return res
		}

		panic("unknown type: " + t.TypeName)

	case parser.ArrayTypeNode:
		itemType := parserTypeToType(t.ItemType)
		return &types.Array{
			Type:     itemType,
			LlvmType: llvmTypes.NewArray(itemType.LLVM(), t.Len),
		}

	case parser.StructTypeNode:
		var structTypes []llvmTypes.Type
		members := make(map[string]types.Type)
		memberIndexes := t.Names

		inverseNamesIndex := make(map[int]string)
		for name, index := range memberIndexes {
			inverseNamesIndex[index] = name
		}

		for i, tt := range t.Types {
			ty := parserTypeToType(tt)

			members[inverseNamesIndex[i]] = ty
			structTypes = append(structTypes, ty.LLVM())

			// memberIndexes[]

			// llvmType := typeNodeToLLVMType(tt)
			// structTypes = append(structTypes, llvmType)

			// elements = append(elements, structElement{
			// 	datatype: datatype{},
			// 	index:    i,
			// })

			// elements[inverseNamesIndex[i]] = structElement{
			// 	index: i,
			// 	datatype: datatype{
			// 		Type:         llvmType,
			// 		baseType:     tt.Type(),
			// 		pointerLevel: 0,
			// 	},
			// }
		}

		return &types.Struct{
			SourceName:    t.Type(),
			Members:       members,
			MemberIndexes: memberIndexes,
			Type:          llvmTypes.NewStruct(structTypes...),
		}

	case parser.SliceTypeNode:
		itemType := parserTypeToType(t.ItemType)
		return &types.Slice{
			Type:     itemType,
			LlvmType: internal.Slice(itemType.LLVM()),
		}
	}

	panic(fmt.Sprintf("unknown typeNode: %T", typeNode))
}
