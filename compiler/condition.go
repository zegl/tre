package compiler

import (
	"github.com/llir/llvm/ir"
	"github.com/zegl/tre/parser"
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
