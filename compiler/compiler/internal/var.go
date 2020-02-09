package internal

import (
	"github.com/llir/llvm/ir"
	llvmValue "github.com/llir/llvm/ir/value"
	"github.com/zegl/tre/compiler/compiler/internal/pointer"
	"github.com/zegl/tre/compiler/compiler/value"
)

func LoadIfVariable(block *ir.Block, val value.Value) llvmValue.Value {
	if val.IsVariable {
		return block.NewLoad(pointer.ElemType(val.Value), val.Value)
	}
	return val.Value
}
