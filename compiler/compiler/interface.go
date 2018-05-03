package compiler

import (
	"log"

	"github.com/llir/llvm/ir/constant"
	llvmTypes "github.com/llir/llvm/ir/types"
	"github.com/zegl/tre/compiler/compiler/internal"
	"github.com/zegl/tre/compiler/compiler/types"
	"github.com/zegl/tre/compiler/compiler/value"
)

func (c *Compiler) valueToInterfaceValue(v value.Value, targetType types.Type) value.Value {
	_, isInterface := targetType.(*types.Interface)

	// Don't do anything if the target is not an interface
	if !isInterface {
		log.Printf("valueToInterfaceValue: not converting to interface")
		return v
	}

	log.Printf("valueToInterfaceValue: converting to interface")

	val := v.Value

	// Convert to pointer variable
	if !v.IsVariable {
		ptrAlloca := c.contextBlock.NewAlloca(v.Type.LLVM())
		c.contextBlock.NewStore(val, ptrAlloca)
		val = ptrAlloca
	}

	ifaceStruct := c.contextBlock.NewAlloca(internal.Interface())

	dataPtr := c.contextBlock.NewGetElementPtr(ifaceStruct, constant.NewInt(0, i32.LLVM()), constant.NewInt(0, i32.LLVM()))
	bitcastedVal := c.contextBlock.NewBitCast(val, llvmTypes.NewPointer(llvmTypes.I8))
	c.contextBlock.NewStore(bitcastedVal, dataPtr)

	dataTypePtr := c.contextBlock.NewGetElementPtr(ifaceStruct, constant.NewInt(0, i32.LLVM()), constant.NewInt(1, i32.LLVM()))

	backingTypID, ok := typeID[v.Type.Name()]
	if !ok {
		panic(v.Type.Name() + " has no typeID")
	}

	c.contextBlock.NewStore(constant.NewInt(backingTypID, i32.LLVM()), dataTypePtr)

	// val = c.contextBlock.NewLoad(ifaceStruct)

	return value.Value{
		Type:       &types.Interface{},
		Value:      ifaceStruct,
		IsVariable: true,
	}
}
