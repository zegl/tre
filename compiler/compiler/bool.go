package compiler

import (
	"github.com/llir/llvm/ir/constant"
	"github.com/zegl/tre/compiler/compiler/types"
	"github.com/zegl/tre/compiler/compiler/value"
	"github.com/zegl/tre/compiler/parser"
)

func (c *Compiler) compileNegateBoolNode(v parser.NegateNode) value.Value {

	val := c.compileValue(v.Item)
	loadedVal := c.contextBlock.NewLoad(val.Value)

	return value.Value{
		Type:       types.Bool,
		Value:      c.contextBlock.NewXor(loadedVal, constant.NewInt(1, types.Bool.LLVM())),
		IsVariable: false,
	}
}
