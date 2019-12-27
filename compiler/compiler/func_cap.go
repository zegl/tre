package compiler

import (
	"github.com/zegl/tre/compiler/compiler/internal/pointer"
	"github.com/zegl/tre/compiler/compiler/value"
	"github.com/zegl/tre/compiler/parser"
)

func (c *Compiler) capFuncCall(v *parser.CallNode) value.Value {
	arg := c.compileValue(v.Arguments[0])

	if arg.Type.Name() == "slice" {
		val := arg.Value
		val = c.contextBlock.NewLoad(pointer.ElemType(val), val)

		return value.Value{
			Value:      c.contextBlock.NewExtractValue(val, 1),
			Type:       i64,
			IsVariable: false,
		}
	}

	c.panic(c.contextBlock, "Can not call cap on "+arg.Type.Name())
	return value.Value{}
}
