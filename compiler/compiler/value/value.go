package value

import (
	llvmValue "github.com/llir/llvm/ir/value"
	"github.com/zegl/tre/compiler/compiler/types"
)

type Value struct {
	Type         types.Type
	Value        llvmValue.Value
	PointerLevel uint
}
