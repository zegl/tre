package internal

import (
	"github.com/llir/llvm/ir/types"
)

func String() *types.StructType {
	return types.NewStruct(
		types.I32,                  // String length
		types.NewPointer(types.I8), // Content
	)
}
