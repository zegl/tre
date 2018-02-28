package types

import "github.com/llir/llvm/ir/types"

type PackageInstance struct {
	backingType

	Funcs map[string]*Function
}

func (PackageInstance) LLVM() types.Type {
	panic("PackageInstance has not LLVM")
}

func (PackageInstance) Name() string {
	return "package"
}
