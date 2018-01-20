package internal

import (
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/types"
)

func String() *types.StructType {
	return types.NewStruct(
		types.I32,                  // String length
		types.NewPointer(types.I8), // Content
	)
}

func StringLen(stringType types.Type) *ir.Function {
	param := ir.NewParam("input", stringType)
	res := ir.NewFunction("string_len", types.I32, param)
	block := res.NewBlock("entry")
	block.NewRet(block.NewExtractValue(param, []int64{0}))
	return res
}
