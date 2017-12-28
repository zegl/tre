package compiler

import "github.com/llir/llvm/ir/types"

var typeConvertMap = map[string]types.Type{
	"i64": types.I64,
	"i32": types.I32,
}

func convertTypes(sourceName string) types.Type {
	if t, ok := typeConvertMap[sourceName]; ok {
		return t
	}

	panic("unknown type: " + sourceName)
}
