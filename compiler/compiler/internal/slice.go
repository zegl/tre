package internal

import (
	"github.com/llir/llvm/ir/types"
)

func Slice(itemType types.Type) *types.StructType {
	return types.NewStruct(
		types.I32,                  // Len
		types.I32,                  // Cap
		types.I32,                  // Array Offset
		types.NewPointer(itemType), // Content
	)
}
