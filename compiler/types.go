package compiler

import "github.com/llir/llvm/ir/types"

var typeConvertMap = map[string]types.Type{
	"i8": types.I8,
	"i32": types.I32,
	"i64": types.I64,
}

func typeStringToLLVM(sourceName string) types.Type {
	if t, ok := typeConvertMap[sourceName]; ok {
		return t
	}

	panic("unknown type: " + sourceName)
}
