package strings

import (
	"fmt"

	"github.com/llir/llvm/ir/types"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/value"
)

func Constant(in string) *constant.Array {
	var constants []constant.Constant

	for _, char := range in {
		constants = append(constants, constant.NewInt(int64(char), types.I8))
	}

	// null
	constants = append(constants, constant.NewInt(0, types.I8))

	s := constant.NewArray(constants...)
	s.CharArray = true

	return s
}

func Toi8Ptr(block *ir.BasicBlock, src value.Value) *ir.InstGetElementPtr {
	return block.NewGetElementPtr(src, constant.NewInt(0, types.I64), constant.NewInt(0, types.I64))
}

func TreStringToi8Ptr(block *ir.BasicBlock, src value.Value) *ir.InstExtractValue {
	return block.NewExtractValue(src, []int64{1})
}

var globalStringCounter uint

func NextStringName() string {
	name := fmt.Sprintf("str.%d", globalStringCounter)
	globalStringCounter++
	return name
}
