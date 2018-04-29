package compiler

import (
	"fmt"

	"github.com/llir/llvm/ir/constant"
	"github.com/zegl/tre/compiler/compiler/types"
	"github.com/zegl/tre/compiler/compiler/value"
	"github.com/zegl/tre/compiler/parser"
)

func (c *Compiler) compileStructLoadElementNode(v parser.StructLoadElementNode) value.Value {
	src := c.compileValue(v.Struct)

	if packageRef, ok := src.Type.(*types.PackageInstance); ok {
		if f, ok := packageRef.Funcs[v.ElementName]; ok {
			return value.Value{
				Type: f,
			}
		}

		panic(fmt.Sprintf("Package %s has no such method %s", packageRef.Name(), v.ElementName))
	}

	if !src.IsVariable {
		// GetElementPtr only works on pointer types, and we don't have a pointer to our object.
		// Allocate it and use the pointer instead
		dst := c.contextBlock.NewAlloca(src.Type.LLVM())
		c.contextBlock.NewStore(src.Value, dst)
		src = value.Value{
			Value:      dst,
			Type:       src.Type,
			IsVariable: true,
		}
	}

	// Check if it is a struct member
	if structType, ok := src.Type.(*types.Struct); ok {
		if memberIndex, ok := structType.MemberIndexes[v.ElementName]; ok {
			retVal := c.contextBlock.NewGetElementPtr(src.Value, constant.NewInt(0, i32.LLVM()), constant.NewInt(int64(memberIndex), i32.LLVM()))
			return value.Value{
				Type:       structType.Members[v.ElementName],
				Value:      retVal,
				IsVariable: true,
			}
		}
	}

	// Check if it's a method
	if method, ok := src.Type.GetMethod(v.ElementName); ok {
		return value.Value{
			Type:       method,
			Value:      src.Value,
			IsVariable: false,
		}
	}

	panic(fmt.Sprintf("%T internal error: no such type map indexing: %s", src, v.ElementName))
}
