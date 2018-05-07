package compiler

import (
	"github.com/llir/llvm/ir/constant"
	llvmTypes "github.com/llir/llvm/ir/types"
	"github.com/zegl/tre/compiler/compiler/types"
	"github.com/zegl/tre/compiler/compiler/value"
)

func (c *Compiler) valueToInterfaceValue(v value.Value, targetType types.Type) value.Value {

	// Don't do anything if the target is not an interface
	if _, isInterface := targetType.(*types.Interface); !isInterface {
		return v
	}

	// Don't do anything if the src already is an interface
	if _, sourceIsInterface := v.Type.(*types.Interface); sourceIsInterface {
		return v
	}

	val := v.Value

	// Convert to pointer variable
	if !v.IsVariable {
		ptrAlloca := c.contextBlock.NewAlloca(v.Type.LLVM())
		c.contextBlock.NewStore(val, ptrAlloca)
		val = ptrAlloca
	}

	ifaceStruct := c.contextBlock.NewAlloca(targetType.LLVM())

	dataPtr := c.contextBlock.NewGetElementPtr(ifaceStruct, constant.NewInt(0, i32.LLVM()), constant.NewInt(0, i32.LLVM()))
	bitcastedVal := c.contextBlock.NewBitCast(val, llvmTypes.NewPointer(llvmTypes.I8))
	c.contextBlock.NewStore(bitcastedVal, dataPtr)

	dataTypePtr := c.contextBlock.NewGetElementPtr(ifaceStruct, constant.NewInt(0, i32.LLVM()), constant.NewInt(1, i32.LLVM()))

	backingTypID := getTypeID(v.Type.Name())
	c.contextBlock.NewStore(constant.NewInt(backingTypID, i32.LLVM()), dataTypePtr)

	// Add methods to the iface table
	// TODO

	return value.Value{
		Type:       targetType,
		Value:      ifaceStruct,
		IsVariable: true,
	}
}
