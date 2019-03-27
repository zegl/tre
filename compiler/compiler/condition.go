package compiler

import (
	"fmt"
	"github.com/zegl/tre/compiler/compiler/name"

	"github.com/zegl/tre/compiler/compiler/types"
	"github.com/zegl/tre/compiler/compiler/value"
	"github.com/zegl/tre/compiler/parser"

	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	llvmTypes "github.com/llir/llvm/ir/types"
	llvmValue "github.com/llir/llvm/ir/value"
)

func getConditionLLVMpred(operator parser.Operator) enum.IPred {
	m := map[parser.Operator]enum.IPred{
		parser.OP_GT:   enum.IPredSGT,
		parser.OP_GTEQ: enum.IPredSGE,
		parser.OP_LT:   enum.IPredSLT,
		parser.OP_LTEQ: enum.IPredSLE,
		parser.OP_EQ:   enum.IPredEQ,
		parser.OP_NEQ:  enum.IPredNE,
	}

	if op, ok := m[operator]; ok {
		return op
	}

	panic("unknown op: " + string(operator))
}

func (c *Compiler) compileOperatorNode(v *parser.OperatorNode) value.Value {
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

	switch leftLLVM.Type().Name() {
	case "string":
		if v.Operator == parser.OP_ADD {
			leftLen := c.contextBlock.NewExtractValue(leftLLVM, 0)
			rightLen := c.contextBlock.NewExtractValue(rightLLVM, 0)
			sumLen := c.contextBlock.NewAdd(leftLen, rightLen)

			backingArray := c.contextBlock.NewAlloca(i8.LLVM())
			backingArray.NElems = sumLen

			// Copy left to new backing array
			c.contextBlock.NewCall(c.externalFuncs.Strcpy.LlvmFunction, backingArray, c.contextBlock.NewExtractValue(leftLLVM, 1))

			// Append right to backing array
			c.contextBlock.NewCall(c.externalFuncs.Strcat.LlvmFunction, backingArray, c.contextBlock.NewExtractValue(rightLLVM, 1))

			alloc := c.contextBlock.NewAlloca(typeConvertMap["string"].LLVM())

			// Save length of the string
			lenItem := c.contextBlock.NewGetElementPtr(alloc, constant.NewInt(llvmTypes.I32, 0), constant.NewInt(llvmTypes.I32, 0))
			c.contextBlock.NewStore(sumLen, lenItem)

			// Save i8* version of string
			strItem := c.contextBlock.NewGetElementPtr(alloc, constant.NewInt(llvmTypes.I32, 0), constant.NewInt(llvmTypes.I32, 1))
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

func (c *Compiler) compileSubNode(v *parser.SubNode) value.Value {
	right := c.compileValue(v.Item)
	rVal := right.Value

	if right.IsVariable {
		rVal = c.contextBlock.NewLoad(rVal)
	}

	res := c.contextBlock.NewSub(
		constant.NewInt(rVal.Type().(*llvmTypes.IntType), 0),
		rVal,
	)

	return value.Value {
		Value: res,
		Type: right.Type,
		IsVariable: false,
	}
}

func (c *Compiler) compileConditionNode(v *parser.ConditionNode) {

	c.pushVariablesStack()

	cond := c.compileOperatorNode(v.Cond)

	afterBlock := c.contextBlock.Parent.NewBlock(name.Block() + "-after")
	trueBlock := c.contextBlock.Parent.NewBlock(name.Block() + "-true")
	falseBlock := afterBlock

	// push afterBlock stack
	c.contextCondAfter = append(c.contextCondAfter, afterBlock)

	if len(v.False) > 0 {
		falseBlock = c.contextBlock.Parent.NewBlock(name.Block() + "-false")
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

	// pop after block stack
	c.contextCondAfter = c.contextCondAfter[0 : len(c.contextCondAfter)-1]

	// set after block to jump to the after block
	if len(c.contextCondAfter) > 0 {
		afterBlock.NewBr(c.contextCondAfter[len(c.contextCondAfter)-1])
	}

	c.popVariablesStack()
}
