package compiler

import (
	"fmt"
	"log"

	"github.com/llir/llvm/ir/constant"
	llvmTypes "github.com/llir/llvm/ir/types"
	llvmValue "github.com/llir/llvm/ir/value"
	"github.com/zegl/tre/compiler/compiler/types"
	"github.com/zegl/tre/compiler/compiler/value"
	"github.com/zegl/tre/compiler/parser"
)

func (c *Compiler) compileStructLoadElementNode(v *parser.StructLoadElementNode) value.Value {
	src := c.compileValue(v.Struct)

	if packageRef, ok := src.Type.(*types.PackageInstance); ok {
		if f, ok := packageRef.Funcs[v.ElementName]; ok {
			return value.Value{
				Type: f,
			}
		}

		panic(fmt.Sprintf("Package %s has no such method %s", packageRef.Name(), v.ElementName))
	}

	// Use this type, or the type behind the pointer
	targetType := src.Type
	var isPointer bool
	var isPointerNonAllocDereference bool
	if pointerType, ok := src.Type.(*types.Pointer); ok {
		targetType = pointerType.Type
		isPointerNonAllocDereference = pointerType.IsNonAllocDereference
		isPointer = true
	}

	if !src.IsVariable && !isPointer {
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
	if structType, ok := targetType.(*types.Struct); ok {
		if memberIndex, ok := structType.MemberIndexes[v.ElementName]; ok {

			val := src.Value

			var retVal llvmValue.Value

			if isPointer && !isPointerNonAllocDereference && structType.IsHeapAllocated {
				val = c.contextBlock.NewLoad(src.Value)
				log.Printf("mem: %+v", val)
				retVal = c.contextBlock.NewGetElementPtr(val, constant.NewInt(0, i32.LLVM()), constant.NewInt(int64(memberIndex), i32.LLVM()))
			} else {
				retVal = c.contextBlock.NewGetElementPtr(val, constant.NewInt(0, i32.LLVM()), constant.NewInt(int64(memberIndex), i32.LLVM()))
			}

			return value.Value{
				Type:       structType.Members[v.ElementName],
				Value:      retVal,
				IsVariable: true,
			}
		}
	}

	// Check if it's a method
	if method, ok := targetType.GetMethod(v.ElementName); ok {
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

func (c *Compiler) compileInitStructWithValues(v *parser.InitializeStructNode) value.Value {
	treType := parserTypeToType(v.Type)

	structType, ok := treType.(*types.Struct)
	if !ok {
		panic("Expected struct type in compileInitStructWithValues")
	}

	var alloc llvmValue.Value

	// Allocate on the heap or on the stack
	if len(c.contextAlloc) > 0 && c.contextAlloc[len(c.contextAlloc)-1].Escapes {
		mallocatedSpaceRaw := c.contextBlock.NewCall(c.externalFuncs.Malloc.LlvmFunction, constant.NewInt(structType.Size(), i64.LLVM()))
		alloc = c.contextBlock.NewBitCast(mallocatedSpaceRaw, llvmTypes.NewPointer(structType.LLVM()))
		structType.IsHeapAllocated = true
	} else {
		alloc = c.contextBlock.NewAlloca(structType.LLVM())
	}

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
		Type:       structType,
		Value:      alloc,
		IsVariable: true,
	}
}
