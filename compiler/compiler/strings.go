package compiler

import (
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/value"
	"github.com/llir/llvm/ir/constant"
	"fmt"
)

func constantString(in string) *constant.Array {
	var constants []constant.Constant

	for _, char := range in {
		constants = append(constants, constant.NewInt(int64(char), i8))
	}

	// null
	constants = append(constants, constant.NewInt(0, i8))

	s := constant.NewArray(constants...)
	s.CharArray = true

	return s
}

func stringToi8Ptr(block *ir.BasicBlock, src value.Value) *ir.InstGetElementPtr {
	return block.NewGetElementPtr(src, constant.NewInt(0, i64), constant.NewInt(0, i64))
}

var globalStringCounter uint

func getNextStringName() string {
	name := fmt.Sprintf("str.%d", globalStringCounter)
	globalStringCounter++
	return name
}
