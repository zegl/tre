package compiler

import (
	"fmt"

	"github.com/zegl/tre/compiler/compiler/internal"

	"github.com/zegl/tre/compiler/compiler/types"
	"github.com/zegl/tre/compiler/parser"

	llvmTypes "github.com/llir/llvm/ir/types"
)

var typeConvertMap = map[string]types.Type{
	"bool":   types.I1,
	"int":    types.I64, // TODO: Size based on arch
	"int8":   types.I8,
	"int16":  types.I16,
	"int32":  types.I32,
	"int64":  types.I64,
	"string": types.String,
}

func parserTypeToType(typeNode parser.TypeNode) types.Type {
	switch t := typeNode.(type) {
	case parser.SingleTypeNode:
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
