package compiler

import (
	"fmt"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	llvmTypes "github.com/llir/llvm/ir/types"
	"github.com/zegl/tre/compiler/compiler/types"
	"github.com/zegl/tre/compiler/compiler/value"
)

func (c *Compiler) valueToInterfaceValue(v value.Value, targetType types.Type) value.Value {

	// Don't do anything if the target is not an interface
	iface, isInterface := targetType.(*types.Interface)
	if !isInterface {
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

	// Create interface jump table if needed
	var funcTableAlloca *ir.InstAlloca
	if len(iface.RequiredMethods) > 0 {
		funcTablePtr := c.contextBlock.NewGetElementPtr(ifaceStruct,
			constant.NewInt(0, i32.LLVM()),
			constant.NewInt(2, i32.LLVM()),
		)
		funcTableAlloca = c.contextBlock.NewAlloca(iface.JumpTable())
		c.contextBlock.NewStore(funcTableAlloca, funcTablePtr)
	}

	// Add methods to the iface table
	for methodIndex, methodName := range iface.SortedRequiredMethods() {
		functionPointer := c.contextBlock.NewGetElementPtr(funcTableAlloca,
			constant.NewInt(0, i32.LLVM()),
			constant.NewInt(int64(methodIndex), i32.LLVM()),
		)

		m, ok := v.Type.GetMethod(methodName)
		if !ok {
			panic(fmt.Sprintf("%s can not be used as %s, is missing %s method", v.Type.Name(), targetType.Name(), methodName))
		}
		c.contextBlock.NewStore(m.Function.JumpFunction, functionPointer)
	}

	return value.Value{
		Type:       targetType,
		Value:      ifaceStruct,
		IsVariable: true,
	}
}
