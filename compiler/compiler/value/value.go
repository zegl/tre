package value

import (
	"github.com/llir/llvm/ir/constant"
	llvmValue "github.com/llir/llvm/ir/value"

	"github.com/zegl/tre/compiler/compiler/types"
)

type Value struct {
	Type  types.Type
	Value llvmValue.Value

	// Is true when Value points to an LLVM Allocated variable, and is false
	// when the value is a constant.
	// This is used to know if a "load" instruction is necessary or not.
	// Pointers are not considered "variables" in this context.
	IsVariable bool

	// Is used when returning multiple types from a function
	// Type is set to MultiValue when this is case, and will also contain the
	// type information
	MultiValues []Value
}

func UntypedConstAs(val Value, context Value) Value {
	switch val.Type.(type) {
	case *types.UntypedConstantNumber:
		if contextInt, ok := context.Type.(*types.Int); ok {
			return Value{
				Type: contextInt,
				Value: &constant.Int{
					Typ: contextInt.Type,
					X:   val.Value.(*constant.Int).X,
				},
			}
		}
		panic("unexpected type in UntypedConstAs")
	default:
		panic("unexpected type in UntypedConstAs")
	}
}
