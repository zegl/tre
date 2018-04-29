package compiler

import (
	"fmt"

	"github.com/zegl/tre/compiler/compiler/internal"
	"github.com/zegl/tre/compiler/compiler/value"

	"github.com/zegl/tre/compiler/compiler/types"
	"github.com/zegl/tre/compiler/parser"

	llvmTypes "github.com/llir/llvm/ir/types"
	llvmValue "github.com/llir/llvm/ir/value"
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

	case parser.InterfaceTypeNode:
		return &types.Interface{}
	}

	panic(fmt.Sprintf("unknown typeNode: %T", typeNode))
}

func (c *Compiler) compileTypeCastNode(v parser.TypeCastNode) value.Value {
	val := c.compileValue(v.Val)

	var current *llvmTypes.IntType
	var ok bool

	current, ok = val.Type.LLVM().(*llvmTypes.IntType)
	if !ok {
		panic("TypeCast origin must be int type")
	}

	targetType := parserTypeToType(v.Type)
	target, ok := targetType.LLVM().(*llvmTypes.IntType)
	if !ok {
		panic("TypeCast target must be int type")
	}

	llvmVal := val.Value
	if val.IsVariable {
		llvmVal = c.contextBlock.NewLoad(llvmVal)
	}

	// Same size, nothing to do here
	if current.Size == target.Size {
		return val
	}

	res := c.contextBlock.NewAlloca(target)

	var changedSize llvmValue.Value

	if current.Size < target.Size {
		changedSize = c.contextBlock.NewSExt(llvmVal, target)
	} else {
		changedSize = c.contextBlock.NewTrunc(llvmVal, target)
	}

	c.contextBlock.NewStore(changedSize, res)

	return value.Value{
		Value:      res,
		Type:       targetType,
		IsVariable: true,
	}
}
