package compiler

import "github.com/llir/llvm/ir/types"

var typeConvertMap = map[string]types.Type{
	"int":   types.I64, // TODO: Size based on arch
	"int8":  types.I8,
	"int32": types.I32,
	"int64": types.I64,
}

// Type Name : Element Name : Index
var typeMapElementNameIndex = map[string]map[string]int{}

func typeStringToLLVM(sourceName string) types.Type {
	if t, ok := typeConvertMap[sourceName]; ok {
		return t
	}

	panic("unknown type: " + sourceName)
}
