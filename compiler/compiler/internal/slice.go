package internal

import (
	"github.com/llir/llvm/ir/types"
)

func Slice(itemType types.Type) *types.StructType {
	return types.NewStruct(
		types.I64,                  // Len
		types.I64,                  // Cap
		types.I64,                  // Array Offset
		types.NewPointer(itemType), // Content
	)
}
