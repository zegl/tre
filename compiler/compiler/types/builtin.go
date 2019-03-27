package types

import "github.com/llir/llvm/ir/types"

var I8 = &Int{Type: types.I8, TypeName: "int8", TypeSize: 8/8}
var I16 = &Int{Type: types.I16, TypeName: "int16", TypeSize: 18/8}
var I32 = &Int{Type: types.I32, TypeName: "int32", TypeSize: 32/8}
var I64 = &Int{Type: types.I64, TypeName: "int64", TypeSize: 64/8}

var Void = &VoidType{}
var String = &StringType{}
var Bool = &BoolType{}
