package types

import "github.com/llir/llvm/ir/types"

var I1 = &Basic{Type: types.I1, TypeName: "bool", TypeSize: 1}
var I8 = &Basic{Type: types.I8, TypeName: "int8", TypeSize: 1}
var I16 = &Basic{Type: types.I16, TypeName: "int16", TypeSize: 2}
var I32 = &Basic{Type: types.I32, TypeName: "int32", TypeSize: 4}
var I64 = &Basic{Type: types.I64, TypeName: "int64", TypeSize: 8}
var Bool = &Basic{Type: types.I1, TypeName: "bool"}
var Void = &Basic{Type: types.Void, TypeName: "void"}
var String = &StringType{}

var _ Type = I8
