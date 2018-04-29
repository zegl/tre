package value

import (
	llvmValue "github.com/llir/llvm/ir/value"
	"github.com/zegl/tre/compiler/compiler/types"
)

type Value struct {
	Type  types.Type
	Value llvmValue.Value

	// Is true when Value points to an LLVM Allocated variable, and is false
	// when the value is a constant.
	// This is used to know if a "load" instruction is neccesary or not.
	IsVariable bool
}
