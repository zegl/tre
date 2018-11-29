package strings

import (
	"fmt"

	"github.com/llir/llvm/ir/types"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/value"
)

func Constant(in string) *constant.CharArray {
	return constant.NewCharArray(append([]byte(in), 0)) 
}

func Toi8Ptr(block *ir.BasicBlock, src value.Value) *ir.InstGetElementPtr {
	return block.NewGetElementPtr(src, constant.NewInt(types.I64, 0), constant.NewInt(types.I64, 0))
}

func TreStringToi8Ptr(block *ir.BasicBlock, src value.Value) *ir.InstExtractValue {
	return block.NewExtractValue(src, 1)
}

var globalStringCounter uint

func NextStringName() string {
	name := fmt.Sprintf("str.%d", globalStringCounter)
	globalStringCounter++
	return name
}
