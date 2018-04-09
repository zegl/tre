package compiler

import (
	"github.com/llir/llvm/ir/constant"
	llvmTypes "github.com/llir/llvm/ir/types"
	"github.com/zegl/tre/compiler/compiler/value"
	"github.com/zegl/tre/compiler/parser"
)

func (c *Compiler) lenFuncCall(v parser.CallNode) value.Value {
	arg := c.compileValue(v.Arguments[0])

	if arg.Type.Name() == "string" {
		f := c.funcByName("len_string")

		val := arg.Value
		if arg.PointerLevel > 0 {
			val = c.contextBlock.NewLoad(val)
		}

		return value.Value{
			Value:        c.contextBlock.NewCall(f.LlvmFunction, val),
			Type:         f.ReturnType,
			PointerLevel: 0,
		}
	}

	if arg.Type.Name() == "array" {
		if ptrType, ok := arg.Value.Type().(*llvmTypes.PointerType); ok {
			if arrayType, ok := ptrType.Elem.(*llvmTypes.ArrayType); ok {
				return value.Value{
					Value:        constant.NewInt(arrayType.Len, i64.LLVM()),
					Type:         i64,
					PointerLevel: 0,
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
			Value:        c.contextBlock.NewExtractValue(val, []int64{0}),
			Type:         i64,
			PointerLevel: 0,
		}
	}

	c.panic(c.contextBlock, "Can not call len on "+arg.Type.Name())
	return value.Value{}
}
