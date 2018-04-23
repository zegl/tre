package compiler

import (
	"github.com/zegl/tre/compiler/compiler/types"
	"github.com/zegl/tre/compiler/compiler/value"
	"github.com/zegl/tre/compiler/parser"

	"github.com/llir/llvm/ir"
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

func (c *Compiler) compileCondition(v parser.OperatorNode) value.Value {
	left := c.compileValue(v.Left)
	right := c.compileValue(v.Right)

	leftVal := left.Value
	rightVal := right.Value

	if left.PointerLevel > 0 {
		leftVal = c.contextBlock.NewLoad(leftVal)
	}

	if right.PointerLevel > 0 {
		rightVal = c.contextBlock.NewLoad(rightVal)
	}

	return value.Value{
		Type:         types.Bool,
		Value:        c.contextBlock.NewICmp(getConditionLLVMpred(v.Operator), leftVal, rightVal),
		PointerLevel: 0,
	}
}
