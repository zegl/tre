package compiler

import (
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	irTypes "github.com/llir/llvm/ir/types"
	llvmValue "github.com/llir/llvm/ir/value"

	"github.com/zegl/tre/compiler/compiler/internal"
	"github.com/zegl/tre/compiler/compiler/internal/pointer"
	"github.com/zegl/tre/compiler/compiler/name"
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

	if v.IsConst {
		c.compileAllocConstNode(v)
		return
	}

	// Allocate from type
	if len(v.Val) == 0 && v.Type != nil {
		treType := c.parserTypeToType(v.Type)

		var val llvmValue.Value
		var block *ir.Block

		// Package level variables
		if c.contextBlock == nil {
			globType := treType.LLVM()
			glob := c.module.NewGlobal(name.Var(v.Name[0]), globType)
			glob.Init = constant.NewZeroInitializer(globType)

			c.currentPackage.DefinePkgVar(v.Name[0], value.Value{
				Value:      glob,
				Type:       treType,
				IsVariable: true,
			})

			val = glob
			block = c.initGlobalsFunc.Blocks[0]
		} else {
			alloc := c.contextBlock.NewAlloca(treType.LLVM())
			alloc.SetName(name.Var(v.Name[0]))

			c.setVar(v.Name[0], value.Value{
				Value:      alloc,
				Type:       treType,
				IsVariable: true,
			})

			val = alloc
			block = c.contextBlock
		}

		// Set to zero values
		// TODO: Make slices less special
		if sliceType, ok := treType.(*types.Slice); ok {
			sliceType.SliceZero(block, c.externalFuncs.Malloc.Value.(llvmValue.Named), 2, val)
		} else {
			treType.Zero(block, val)
		}

		return
	}

	for valIndex, valNode := range v.Val {
		// Allocate from value
		val := c.compileValue(valNode)

		if _, ok := val.Type.(*types.MultiValue); ok {
			if len(v.Name) != len(val.MultiValues) {
				panic("Variable count on left and right side does not match")
			}

			// Is currently expecting that the variables are already allocated in this block.
			// Will only add the vars to the map of variables
			for i, multiVal := range val.MultiValues {
				c.setVar(v.Name[i], multiVal)
			}

			return
		}

		// Single variable allocation
		llvmVal := val.Value

		// Non-allocation needed pointers
		if ptrVal, ok := val.Type.(*types.Pointer); ok && ptrVal.IsNonAllocDereference {
			c.setVar(v.Name[valIndex], value.Value{
				Type:       val.Type,
				Value:      llvmVal,
				IsVariable: false,
			})
			return
		}

		// Non-allocation needed structs
		if structVal, ok := val.Type.(*types.Struct); ok && structVal.IsHeapAllocated {
			c.setVar(v.Name[valIndex], value.Value{
				Type:       val.Type,
				Value:      llvmVal,
				IsVariable: true,
			})
			return
		}

		if val.IsVariable {
			llvmVal = c.contextBlock.NewLoad(pointer.ElemType(llvmVal), llvmVal)
		}

		alloc := c.contextBlock.NewAlloca(llvmVal.Type())
		alloc.SetName(name.Var(v.Name[valIndex]))
		c.contextBlock.NewStore(llvmVal, alloc)

		c.setVar(v.Name[valIndex], value.Value{
			Type:       val.Type,
			Value:      alloc,
			IsVariable: true,
		})
	}

	return
}

func (c *Compiler) compileAllocConstNode(v *parser.AllocNode) {
	for i, varName := range v.Name {
		cnst := v.Val[i].(*parser.ConstantNode)
		c.setVar(varName, value.Value{
			Type:  &types.UntypedConstantNumber{},
			Value: constant.NewInt(i64.LLVM().(*irTypes.IntType), cnst.Value),
		})
	}
}

func (c *Compiler) compileAssignNode(v *parser.AssignNode) {
	tmpStores := make([]llvmValue.Value, len(v.Target))
	realTargets := make([]value.Value, len(v.Target))

	// Skip temporary variables if we're assigning to one single var
	if len(v.Target) == 1 {
		dst := c.compileValue(v.Target[0])
		if !dst.IsVariable {
			compilePanic("Can only assign to variable")
		}
		s := c.compileSingleAssign(dst.Type, dst, v.Val[0])
		c.contextBlock.NewStore(s, dst.Value)
		return
	}

	for i := range v.Target {
		target := v.Target[i]

		dst := c.compileValue(target)

		// Allocate a temporary storage
		llvmType := dst.Value.Type()

		if dst.IsVariable {
			p := llvmType.(*irTypes.PointerType)
			llvmType = p.ElemType
		}

		singleAssignVal := c.compileSingleAssign(dst.Type, dst, v.Val[i])

		tmpStore := c.contextBlock.NewAlloca(llvmType)
		c.contextBlock.NewStore(singleAssignVal, tmpStore)
		tmpStores[i] = tmpStore
		realTargets[i] = dst
	}

	for i := range v.Target {
		x := c.contextBlock.NewLoad(pointer.ElemType(tmpStores[i]), tmpStores[i])
		c.contextBlock.NewStore(x, realTargets[i].Value)
	}
}

func (c *Compiler) compileSingleAssign(temporaryDst types.Type, realDst value.Value, val parser.Node) llvmValue.Value {
	// Push assign type stack
	// Can be used later when evaluating integer constants
	// Is also used by append()
	c.contextAssignDest = append(c.contextAssignDest, realDst)

	// Allocate from value
	comVal := c.compileValue(val)

	// Cast to interface if needed
	comVal = c.valueToInterfaceValue(comVal, temporaryDst)
	llvmV := internal.LoadIfVariable(c.contextBlock, comVal)

	// Pop assigng type stack
	c.contextAssignDest = c.contextAssignDest[0 : len(c.contextAssignDest)-1]

	return llvmV
}
