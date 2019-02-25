package compiler

import (
	"fmt"

	"github.com/llir/llvm/ir/constant"
	llvmTypes "github.com/llir/llvm/ir/types"
	"github.com/zegl/tre/compiler/compiler/value"
	"github.com/zegl/tre/compiler/parser"
)

func (c *Compiler) lenFuncCall(v *parser.CallNode) value.Value {
	arg := c.compileValue(v.Arguments[0])

	if arg.Type.Name() == "string" {
		f, ok := c.funcByName("len_string")
		if !ok {
			panic("could not find len_string func")
		}

		val := arg.Value
		if arg.IsVariable {
			val = c.contextBlock.NewLoad(val)
		}

		return value.Value{
			Value:      c.contextBlock.NewCall(f.LlvmFunction, val),
			Type:       f.LlvmReturnType,
			IsVariable: false,
		}
	}

	if arg.Type.Name() == "array" {
		if ptrType, ok := arg.Value.Type().(*llvmTypes.PointerType); ok {
			if arrayType, ok := ptrType.ElemType.(*llvmTypes.ArrayType); ok {
				return value.Value{
					Value:      constant.NewInt(llvmTypes.I32, int64(arrayType.Len)),
					Type:       i32,
					IsVariable: false,
				}
			}
		}
	}

	if arg.Type.Name() == "slice" {
		val := arg.Value
		val = c.contextBlock.NewLoad(val)

		// TODO: Why is a double load needed?
		if _, ok := val.Type().(*llvmTypes.PointerType); ok {
			val = c.contextBlock.NewLoad(val)
		}

		return value.Value{
			Value:      c.contextBlock.NewExtractValue(val, 0),
			Type:       i32,
			IsVariable: false,
		}
	}

	panic(fmt.Sprintf("Can not call len() on type %s (%+v)", arg.Type.Name(), v.Arguments[0]))
}
