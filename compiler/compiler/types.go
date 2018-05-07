package compiler

import (
	"fmt"

	"github.com/zegl/tre/compiler/compiler/internal"
	"github.com/zegl/tre/compiler/compiler/value"

	"github.com/zegl/tre/compiler/compiler/types"
	"github.com/zegl/tre/compiler/parser"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	llvmTypes "github.com/llir/llvm/ir/types"
	llvmValue "github.com/llir/llvm/ir/value"
)

var typeConvertMap = map[string]types.Type{
	"bool":   types.Bool,
	"int":    types.I64, // TODO: Size based on arch
	"int8":   types.I8,
	"int16":  types.I16,
	"int32":  types.I32,
	"int64":  types.I64,
	"string": types.String,
}

// Is used in interfaces to keep track of the backing data type
var typeIDs = map[string]int64{}
var nextTypeID int64

func getTypeID(typeName string) int64 {
	if id, ok := typeIDs[typeName]; ok {
		return id
	}

	nextTypeID++
	typeIDs[typeName] = nextTypeID
	return nextTypeID
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
		requiredMethods := make(map[string]types.InterfaceMethod)

		for name, def := range t.Methods {
			ifaceMethod := types.InterfaceMethod{
				ArgumentTypes: make([]types.Type, 0),
				ReturnTypes:   make([]types.Type, 0),
			}

			for _, arg := range def.ArgumentTypes {
				ifaceMethod.ArgumentTypes = append(ifaceMethod.ArgumentTypes, parserTypeToType(arg))
			}
			for _, ret := range def.ReturnTypes {
				ifaceMethod.ReturnTypes = append(ifaceMethod.ReturnTypes, parserTypeToType(ret))
			}

			requiredMethods[name] = ifaceMethod
		}

		return &types.Interface{RequiredMethods: requiredMethods}
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

func (c *Compiler) compileTypeCastInterfaceNode(v parser.TypeCastInterfaceNode) value.Value {
	tryCastToType := parserTypeToType(v.Type)

	// Allocate the OK variable
	okVal := c.contextBlock.NewAlloca(types.Bool.LLVM())
	types.Bool.Zero(c.contextBlock, okVal)
	okVal.SetName(getVarName("ok"))

	resCastedVal := c.contextBlock.NewAlloca(tryCastToType.LLVM())
	tryCastToType.Zero(c.contextBlock, resCastedVal)
	resCastedVal.SetName(getVarName("rescastedval"))

	interfaceVal := c.compileValue(v.Item)

	interfaceDataType := c.contextBlock.NewGetElementPtr(interfaceVal.Value, constant.NewInt(0, i32.LLVM()), constant.NewInt(1, i32.LLVM()))
	loadedInterfaceDataType := c.contextBlock.NewLoad(interfaceDataType)

	trueBlock := c.contextBlock.Parent.NewBlock(getBlockName() + "-was-correct-type")
	falseBlock := c.contextBlock.Parent.NewBlock(getBlockName() + "-was-other-type")
	afterBlock := c.contextBlock.Parent.NewBlock(getBlockName() + "-after-type-check")

	trueBlock.NewBr(afterBlock)
	falseBlock.NewBr(afterBlock)

	backingTypeID := getTypeID(tryCastToType.Name())
	cmp := c.contextBlock.NewICmp(ir.IntEQ, loadedInterfaceDataType, constant.NewInt(backingTypeID, i32.LLVM()))
	c.contextBlock.NewCondBr(cmp, trueBlock, falseBlock)

	trueBlock.NewStore(constant.NewInt(1, types.Bool.LLVM()), okVal)

	backingDataPtr := trueBlock.NewGetElementPtr(interfaceVal.Value, constant.NewInt(0, i32.LLVM()), constant.NewInt(0, i32.LLVM()))
	loadedBackingDataPtr := trueBlock.NewLoad(backingDataPtr)
	casted := trueBlock.NewBitCast(loadedBackingDataPtr, llvmTypes.NewPointer(tryCastToType.LLVM()))
	loadedCasted := trueBlock.NewLoad(casted)
	trueBlock.NewStore(loadedCasted, resCastedVal)

	c.contextBlock = afterBlock

	return value.Value{
		Type: &types.MultiValue{
			Types: []types.Type{
				tryCastToType,
				types.Bool,
			},
		},
		MultiValues: []value.Value{
			value.Value{Type: tryCastToType, Value: resCastedVal, IsVariable: true},
			value.Value{Type: types.Bool, Value: okVal, IsVariable: true},
		},
	}
}
