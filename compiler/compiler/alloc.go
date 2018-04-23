package compiler

import (
	"github.com/llir/llvm/ir"
	"github.com/zegl/tre/compiler/compiler/types"
	"github.com/zegl/tre/compiler/compiler/value"
	"github.com/zegl/tre/compiler/parser"
)

func (c *Compiler) compileAllocNode(v parser.AllocNode) {
	// Allocate from type
	if typeNode, ok := v.Val.(parser.TypeNode); ok {
		treType := parserTypeToType(typeNode)

		var alloc *ir.InstAlloca

		if sliceType, ok := treType.(*types.Slice); ok {
			alloc = sliceType.SliceZero(c.contextBlock, c.externalFuncs["malloc"], 2)
		} else {
			alloc = c.contextBlock.NewAlloca(treType.LLVM())
			treType.Zero(c.contextBlock, alloc)
		}

		alloc.SetName(v.Name)

		c.contextBlockVariables[v.Name] = value.Value{
			Value:        alloc,
			Type:         treType,
			PointerLevel: 1, // This is probably not always correct
		}
		return
	}

	// Allocate from value
	val := c.compileValue(v.Val)
	llvmVal := val.Value

	if val.PointerLevel > 0 {
		llvmVal = c.contextBlock.NewLoad(llvmVal)
	}

	alloc := c.contextBlock.NewAlloca(llvmVal.Type())
	alloc.SetName(v.Name)
	c.contextBlock.NewStore(llvmVal, alloc)

	c.contextBlockVariables[v.Name] = value.Value{
		Type:         val.Type,
		Value:        alloc,
		PointerLevel: 1, // TODO
	}
}

func (c *Compiler) compileAssignNode(v parser.AssignNode) {
	var dst value.Value

	// TODO: Remove AssignNode.Name
	if len(v.Name) > 0 {
		dst = c.varByName(v.Name)
	} else {
		dst = c.compileValue(v.Target)
	}

	llvmDst := dst.Value

	// Allocate from type
	if typeNode, ok := v.Val.(parser.TypeNode); ok {
		if singleTypeNode, ok := typeNode.(parser.SingleTypeNode); ok {
			alloc := c.contextBlock.NewAlloca(parserTypeToType(singleTypeNode).LLVM())
			c.contextBlock.NewStore(c.contextBlock.NewLoad(alloc), llvmDst)
			return
		}

		panic("AssignNode from non TypeNode is not allowed")
	}

	// Allocate from value
	val := c.compileValue(v.Val)
	llvmV := val.Value

	if val.PointerLevel > 0 {
		llvmV = c.contextBlock.NewLoad(llvmV)
	}

	c.contextBlock.NewStore(llvmV, llvmDst)
}
