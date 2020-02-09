package compiler

import (
	"fmt"

	"github.com/llir/llvm/ir/constant"
	llvmTypes "github.com/llir/llvm/ir/types"
	llvmValue "github.com/llir/llvm/ir/value"

	"github.com/zegl/tre/compiler/compiler/internal"
	"github.com/zegl/tre/compiler/compiler/internal/pointer"
	"github.com/zegl/tre/compiler/compiler/value"
	"github.com/zegl/tre/compiler/parser"
)

func (c *Compiler) lenFuncCall(v *parser.CallNode) value.Value {
	arg := c.compileValue(v.Arguments[0])

	if arg.Type.Name() == "string" {
		f, ok := c.packages["global"].GetPkgVar("len_string")
		if !ok {
			panic("could not find len_string func")
		}
		val := internal.LoadIfVariable(c.contextBlock, arg)

		return value.Value{
			Value:      c.contextBlock.NewCall(f.Value.(llvmValue.Named), val),
			Type:       f.Type,
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
		val = c.contextBlock.NewLoad(pointer.ElemType(val), val)

		if _, ok := val.Type().(*llvmTypes.PointerType); ok {
			val = c.contextBlock.NewLoad(pointer.ElemType(val), val)
		}

		return value.Value{
			Value:      c.contextBlock.NewExtractValue(val, 0),
			Type:       i32,
			IsVariable: false,
		}
	}

	panic(fmt.Sprintf("Can not call len() on type %s (%+v)", arg.Type.Name(), v.Arguments[0]))
}
