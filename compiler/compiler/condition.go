package compiler

import (
	"fmt"

	"github.com/zegl/tre/compiler/compiler/types"
	"github.com/zegl/tre/compiler/compiler/value"
	"github.com/zegl/tre/compiler/parser"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	llvmValue "github.com/llir/llvm/ir/value"
)

func getConditionLLVMpred(operator parser.Operator) ir.IntPred {
	m := map[parser.Operator]ir.IntPred{
		parser.OP_GT:   ir.IntSGT,
		parser.OP_GTEQ: ir.IntSGE,
		parser.OP_LT:   ir.IntSLT,
		parser.OP_LTEQ: ir.IntSLE,
		parser.OP_EQ:   ir.IntEQ,
		parser.OP_NEQ:  ir.IntNE,
	}

	if op, ok := m[operator]; ok {
		return op
	}

	panic("unknown op: " + string(operator))
}

func (c *Compiler) compileOperatorNode(v parser.OperatorNode) value.Value {
	left := c.compileValue(v.Left)
	leftLLVM := left.Value

	right := c.compileValue(v.Right)
	rightLLVM := right.Value

	if left.IsVariable {
		leftLLVM = c.contextBlock.NewLoad(leftLLVM)
	}

	if right.IsVariable {
		rightLLVM = c.contextBlock.NewLoad(rightLLVM)
	}

	if !leftLLVM.Type().Equal(rightLLVM.Type()) {
		panic(fmt.Sprintf("Different types in operation: %T and %T (%+v and %+v)", left.Type, right.Type, leftLLVM.Type(), rightLLVM.Type()))
	}

	switch leftLLVM.Type().GetName() {
	case "string":
		if v.Operator == parser.OP_ADD {
			leftLen := c.contextBlock.NewExtractValue(leftLLVM, []int64{0})
			rightLen := c.contextBlock.NewExtractValue(rightLLVM, []int64{0})
			sumLen := c.contextBlock.NewAdd(leftLen, rightLen)

			backingArray := c.contextBlock.NewAlloca(i8.LLVM())
			backingArray.NElems = sumLen

			// Copy left to new backing array
			c.contextBlock.NewCall(c.externalFuncs.Strcpy.LlvmFunction, backingArray, c.contextBlock.NewExtractValue(leftLLVM, []int64{1}))

			// Append right to backing array
			c.contextBlock.NewCall(c.externalFuncs.Strcat.LlvmFunction, backingArray, c.contextBlock.NewExtractValue(rightLLVM, []int64{1}))

			alloc := c.contextBlock.NewAlloca(typeConvertMap["string"].LLVM())

			// Save length of the string
			lenItem := c.contextBlock.NewGetElementPtr(alloc, constant.NewInt(0, i32.LLVM()), constant.NewInt(0, i32.LLVM()))
			c.contextBlock.NewStore(sumLen, lenItem)

			// Save i8* version of string
			strItem := c.contextBlock.NewGetElementPtr(alloc, constant.NewInt(0, i32.LLVM()), constant.NewInt(1, i32.LLVM()))
			c.contextBlock.NewStore(backingArray, strItem)

			return value.Value{
				Value:      c.contextBlock.NewLoad(alloc),
				Type:       types.String,
				IsVariable: false,
			}
		}

		panic("string does not implement operation " + v.Operator)
	}

	var opRes llvmValue.Value

	switch v.Operator {
	case parser.OP_ADD:
		opRes = c.contextBlock.NewAdd(leftLLVM, rightLLVM)
	case parser.OP_SUB:
		opRes = c.contextBlock.NewSub(leftLLVM, rightLLVM)
	case parser.OP_MUL:
		opRes = c.contextBlock.NewMul(leftLLVM, rightLLVM)
	case parser.OP_DIV:
		opRes = c.contextBlock.NewSDiv(leftLLVM, rightLLVM) // SDiv == Signed Division
	default:
		// Boolean operations
		return value.Value{
			Type:       types.Bool,
			Value:      c.contextBlock.NewICmp(getConditionLLVMpred(v.Operator), leftLLVM, rightLLVM),
			IsVariable: false,
		}
	}

	return value.Value{
		Value:      opRes,
		Type:       left.Type,
		IsVariable: false,
	}
}

func (c *Compiler) compileConditionNode(v parser.ConditionNode) {
	cond := c.compileOperatorNode(v.Cond)

	afterBlock := c.contextBlock.Parent.NewBlock(getBlockName() + "-after")
	trueBlock := c.contextBlock.Parent.NewBlock(getBlockName() + "-true")
	falseBlock := afterBlock

	if len(v.False) > 0 {
		falseBlock = c.contextBlock.Parent.NewBlock(getBlockName() + "-false")
	}

	c.contextBlock.NewCondBr(cond.Value, trueBlock, falseBlock)

	c.contextBlock = trueBlock
	c.compile(v.True)

	// Jump to after-block if no terminator has been set (such as a return statement)
	if trueBlock.Term == nil {
		trueBlock.NewBr(afterBlock)
	}

	if len(v.False) > 0 {
		c.contextBlock = falseBlock
		c.compile(v.False)

		// Jump to after-block if no terminator has been set (such as a return statement)
		if falseBlock.Term == nil {
			falseBlock.NewBr(afterBlock)
		}
	}

	c.contextBlock = afterBlock
}
