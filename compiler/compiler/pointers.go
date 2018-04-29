package compiler

import (
	"fmt"

	"github.com/zegl/tre/compiler/compiler/types"
	"github.com/zegl/tre/compiler/compiler/value"
	"github.com/zegl/tre/compiler/parser"

	llvmTypes "github.com/llir/llvm/ir/types"
)

func (c *Compiler) compileGetReferenceNode(v parser.GetReferenceNode) value.Value {
	val := c.compileValue(v.Item)

	newType := val.Type.LLVM()

	// Up our pointer level one step if val.PointerLevel > 0
	// and val is not of type Pointer
	//
	// TODO: Get rid of the whole PointerLevel mess
	if val.IsVariable {
		if _, ok := val.Type.(*types.Pointer); !ok {
			newType = llvmTypes.NewPointer(newType)
		}
	}

	newSrc := c.contextBlock.NewAlloca(newType)
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

func (c *Compiler) compileDereferenceNode(v parser.DereferenceNode) value.Value {
	val := c.compileValue(v.Item)

	if ptrVal, ok := val.Type.(*types.Pointer); ok {
		return value.Value{
			Value:      c.contextBlock.NewLoad(val.Value),
			Type:       ptrVal.Type,
			IsVariable: true,
		}
	}

	panic(fmt.Sprintf("invalid indirect of TODO (type %s)", val.Type))
}
