package compiler

import (
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
	"github.com/zegl/tre/compiler/parser"
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

func (c *compiler) compileCondition(v parser.OperatorNode) *ir.InstICmp {
	leftVal := c.compileValue(v.Left)
	rightVal := c.compileValue(v.Right)

	for _, val := range []*value.Value{&leftVal, &rightVal} {
		if !loadNeeded(*val) {
			continue
		}

		// Allocate new variable
		// TODO: Is this step needed?
		var newVal *ir.InstAlloca

		if t, valIsPtr := (*val).Type().(*types.PointerType); valIsPtr {
			newVal = c.contextBlock.NewAlloca(t.Elem)
		} else {
			newVal = c.contextBlock.NewAlloca((*val).Type())
		}

		c.contextBlock.NewStore(c.contextBlock.NewLoad(*val), newVal)
		*val = c.contextBlock.NewLoad(newVal)
	}

	return c.contextBlock.NewICmp(getConditionLLVMpred(v.Operator), leftVal, rightVal)
}
