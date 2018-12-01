package compiler

import (
	"github.com/llir/llvm/ir"
	"github.com/zegl/tre/compiler/compiler/types"
	"github.com/zegl/tre/compiler/compiler/value"
	"github.com/zegl/tre/compiler/parser"
)

func (c *Compiler) compileAllocNode(v *parser.AllocNode) {

	// Push and pop alloc stack
	c.contextAlloc = append(c.contextAlloc, v)
	defer func() {
		c.contextAlloc = c.contextAlloc[0 : len(c.contextAlloc)-1]
	}()

	// Allocate from type
	if typeNode, ok := v.Val.(parser.TypeNode); ok {
		treType := parserTypeToType(typeNode)

		var alloc *ir.InstAlloca

		if sliceType, ok := treType.(*types.Slice); ok {
			alloc = sliceType.SliceZero(c.contextBlock, c.externalFuncs.Malloc.LlvmFunction, 2)
		} else {
			alloc = c.contextBlock.NewAlloca(treType.LLVM())
			treType.Zero(c.contextBlock, alloc)
		}

		alloc.SetName(v.Name)

		c.setVar(v.Name, value.Value{
			Value:      alloc,
			Type:       treType,
			IsVariable: true,
		})
		return
	}

	// Allocate from value
	val := c.compileValue(v.Val)

	if _, ok := val.Type.(*types.MultiValue); ok {
		if len(v.MultiNames.Names) != len(val.MultiValues) {
			panic("Variable count on left and right side does not match")
		}

		// Is currently expecting that the variables are already allocated in this block.
		// Will only add the vars to the map of variables
		for i, multiVal := range val.MultiValues {
			c.setVar(v.MultiNames.Names[i].Name, multiVal)
		}

		return
	}

	// Single variable allocation
	llvmVal := val.Value

	// Non-allocation needed pointers
	if ptrVal, ok := val.Type.(*types.Pointer); ok && ptrVal.IsNonAllocDereference {
		c.setVar(v.Name, value.Value{
			Type:       val.Type,
			Value:      llvmVal,
			IsVariable: false,
		})
		return
	}

	// Non-allocation needed structs
	if structVal, ok := val.Type.(*types.Struct); ok && structVal.IsHeapAllocated {
		c.setVar(v.Name, value.Value{
			Type:       val.Type,
			Value:      llvmVal,
			IsVariable: true,
		})
		return
	}

	if val.IsVariable {
		llvmVal = c.contextBlock.NewLoad(llvmVal)
	}

	alloc := c.contextBlock.NewAlloca(llvmVal.Type())
	alloc.SetName(getVarName(v.Name))
	c.contextBlock.NewStore(llvmVal, alloc)

	c.setVar(v.Name, value.Value{
		Type:       val.Type,
		Value:      alloc,
		IsVariable: true,
	})

	return
}

func (c *Compiler) compileAssignNode(v *parser.AssignNode) {
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
		if singleTypeNode, ok := typeNode.(*parser.SingleTypeNode); ok {
			alloc := c.contextBlock.NewAlloca(parserTypeToType(singleTypeNode).LLVM())
			c.contextBlock.NewStore(c.contextBlock.NewLoad(alloc), llvmDst)
			return
		}

		panic("AssignNode from non TypeNode is not allowed")
	}

	// Push assign type stack
	// Can be used later when evaluating integer constants
	// Is also used by append()
	c.contextAssignDest = append(c.contextAssignDest, dst)

	// Allocate from value
	val := c.compileValue(v.Val)

	// Cast to interface if needed
	val = c.valueToInterfaceValue(val, dst.Type)

	llvmV := val.Value

	if val.IsVariable {
		llvmV = c.contextBlock.NewLoad(llvmV)
	}

	c.contextBlock.NewStore(llvmV, llvmDst)

	// Pop assigng type stack
	c.contextAssignDest = c.contextAssignDest[0 : len(c.contextAssignDest)-1]
}
