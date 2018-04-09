package compiler

import (
	llvmTypes "github.com/llir/llvm/ir/types"
	"github.com/zegl/tre/compiler/compiler/value"
	"github.com/zegl/tre/compiler/parser"
)

func (c *Compiler) capFuncCall(v parser.CallNode) value.Value {
	arg := c.compileValue(v.Arguments[0])

	if arg.Type.Name() == "slice" {
		val := arg.Value
		val = c.contextBlock.NewLoad(val)

		// TODO: Why is a double load needed?
		if _, ok := val.Type().(*llvmTypes.PointerType); ok {
			val = c.contextBlock.NewLoad(val)
		}

		return value.Value{
			Value:        c.contextBlock.NewExtractValue(val, []int64{1}),
			Type:         i64,
			PointerLevel: 0,
		}
	}

	c.panic(c.contextBlock, "Can not call cap on "+arg.Type.Name())
	return value.Value{}
}
