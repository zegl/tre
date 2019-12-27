package compiler

import (
	"github.com/llir/llvm/ir/constant"
	llvmTypes "github.com/llir/llvm/ir/types"
	"github.com/zegl/tre/compiler/compiler/internal/pointer"
	"github.com/zegl/tre/compiler/compiler/types"
	"github.com/zegl/tre/compiler/compiler/value"
	"github.com/zegl/tre/compiler/parser"
)

func (c *Compiler) compileNegateBoolNode(v *parser.NegateNode) value.Value {
	val := c.compileValue(v.Item)
	loadedVal := c.contextBlock.NewLoad(pointer.ElemType(val.Value), val.Value)

	return value.Value{
		Type:       types.Bool,
		Value:      c.contextBlock.NewXor(loadedVal, constant.NewInt(llvmTypes.I1, 1)),
		IsVariable: false,
	}
}
