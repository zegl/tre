package internal

import (
	"github.com/llir/llvm/ir/types"
)

func Interface() *types.StructType {
	return types.NewStruct(
		// Pointer to the backing data
		types.NewPointer(types.I8),

		// Backing data type
		types.I32,
	)
}
