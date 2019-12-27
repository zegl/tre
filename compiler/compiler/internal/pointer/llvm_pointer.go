package pointer

import (
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

func ElemType(src value.Value) types.Type {
	return src.Type().(*types.PointerType).ElemType
}
