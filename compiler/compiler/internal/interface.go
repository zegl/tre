package internal

import (
	"github.com/llir/llvm/ir/types"
)

func Interface() *types.StructType {
	return types.NewStruct(
		// TODO: Type information
		types.NewPointer(types.I8), // Backing data
	)
}
