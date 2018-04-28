package types

import "github.com/llir/llvm/ir/types"

var I1 = &Int{Type: types.I1, TypeName: "bool", TypeSize: 1}
var I8 = &Int{Type: types.I8, TypeName: "int8", TypeSize: 1}
var I16 = &Int{Type: types.I16, TypeName: "int16", TypeSize: 2}
var I32 = &Int{Type: types.I32, TypeName: "int32", TypeSize: 4}
var I64 = &Int{Type: types.I64, TypeName: "int64", TypeSize: 8}

var Void = &VoidType{}
var String = &StringType{}
var Bool = &BoolType{}
