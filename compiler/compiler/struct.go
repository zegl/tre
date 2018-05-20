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

	// Check if it's a interface method
	if iface, ok := src.Type.(*types.Interface); ok {
		if ifaceMethod, ok := iface.RequiredMethods[v.ElementName]; ok {
			// Find method index
			// TODO: This can be much smarter
			var methodIndex int64
			for i, name := range iface.SortedRequiredMethods() {
				if name == v.ElementName {
					methodIndex = int64(i)
					break
				}
			}

			// Load jump function
			jumpTable := c.contextBlock.NewGetElementPtr(src.Value, constant.NewInt(0, i32.LLVM()), constant.NewInt(2, i32.LLVM()))
			jumpLoad := c.contextBlock.NewLoad(jumpTable)
			jumpFunc := c.contextBlock.NewGetElementPtr(jumpLoad, constant.NewInt(0, i32.LLVM()), constant.NewInt(methodIndex, i32.LLVM()))
			jumpFuncLoad := c.contextBlock.NewLoad(jumpFunc)

			// Set jump function
			ifaceMethod.LlvmJumpFunction = jumpFuncLoad

			return value.Value{
				Type:       &ifaceMethod,
				Value:      src.Value,
				IsVariable: false,
			}
		}
	}

	panic(fmt.Sprintf("%T internal error: no such type map indexing: %s", src, v.ElementName))
}

func (c *Compiler) compileInitStructWithValues(v parser.InitializeStructNode) value.Value {
	treType := parserTypeToType(v.Type)

	structType, ok := treType.(*types.Struct)
	if !ok {
		panic("Expected struct type in compileInitStructWithValues")
	}

	alloc := c.contextBlock.NewAlloca(treType.LLVM())
	treType.Zero(c.contextBlock, alloc)

	for key, val := range v.Items {
		keyIndex, ok := structType.MemberIndexes[key]
		if !ok {
			panic("Unknown struct key: " + key)
		}

		itemPtr := c.contextBlock.NewGetElementPtr(alloc, constant.NewInt(0, i32.LLVM()), constant.NewInt(int64(keyIndex), i32.LLVM()))
		itemPtr.SetName(getVarName(key))

		compiledVal := c.compileValue(val)

		c.contextBlock.NewStore(compiledVal.Value, itemPtr)
	}

	return value.Value{
		Type:       treType,
		Value:      alloc,
		IsVariable: true,
	}
}
