package internal

import (
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/types"
)

func String() *types.StructType {
	return types.NewStruct(
		types.I64,                  // String length
		types.NewPointer(types.I8), // Content
	)
}

func StringLen(stringType types.Type) *ir.Func {
	param := ir.NewParam("input", stringType)
	res := ir.NewFunc("string_len", types.I64, param)
	block := res.NewBlock("entry")
	block.NewRet(block.NewExtractValue(param, 0))
	return res
}
