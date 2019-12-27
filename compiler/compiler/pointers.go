package compiler

import (
	"fmt"
	"github.com/zegl/tre/compiler/compiler/internal/pointer"
	"github.com/zegl/tre/compiler/compiler/name"

	"github.com/zegl/tre/compiler/compiler/types"
	"github.com/zegl/tre/compiler/compiler/value"
	"github.com/zegl/tre/compiler/parser"
)

func (c *Compiler) compileGetReferenceNode(v *parser.GetReferenceNode) value.Value {
	val := c.compileValue(v.Item)

	// Case where allocation is not necessary, as all LLVM values are pointers by default
	if val.IsVariable {
		if _, ok := val.Type.(*types.Pointer); !ok {
			return value.Value{
				Type: &types.Pointer{
					Type:                  val.Type,
					IsNonAllocDereference: true,
				},
				Value:      val.Value,
				IsVariable: false,
			}
		}

		if structType, ok := val.Type.(*types.Struct); ok && structType.IsHeapAllocated {
			return value.Value{
				Type: &types.Pointer{
					Type:                  val.Type,
					IsNonAllocDereference: true,
				},
				Value:      val.Value,
				IsVariable: false,
			}
		}
	}

	// One extra allocation is neccesary
	newSrc := c.contextBlock.NewAlloca(val.Type.LLVM())
	newSrc.SetName(name.Var("reference-alloca"))
	c.contextBlock.NewStore(val.Value, newSrc)

	return value.Value{
		Type: &types.Pointer{
			Type:     val.Type,
			LlvmType: newSrc.Type(),
		},
		Value:      newSrc,
		IsVariable: true,
	}
}

func (c *Compiler) compileDereferenceNode(v *parser.DereferenceNode) value.Value {
	val := c.compileValue(v.Item)

	if ptrVal, ok := val.Type.(*types.Pointer); ok {
		if ptrVal.IsNonAllocDereference {
			return value.Value{
				Value:      val.Value,
				Type:       ptrVal.Type,
				IsVariable: true,
			}
		}

		return value.Value{
			Value:      c.contextBlock.NewLoad(pointer.ElemType(val.Value), val.Value),
			Type:       ptrVal.Type,
			IsVariable: false,
		}
	}

	panic(fmt.Sprintf("invalid indirect of TODO (type %s)", val.Type))
}
