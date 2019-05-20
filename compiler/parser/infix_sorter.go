package parser

var infixPrioMap = map[Operator]int{
	OP_ADD: 10,
	OP_SUB: 10,

	OP_MUL: 20,
	OP_DIV: 20,
}

func infixPrio(input Operator) int {
	if prio, ok := infixPrioMap[input]; ok {
		return prio
	}

	panic("unknown infixPrio: " + string(input))
}

// sortInfix maxes sure that the infix operators are applied in the correct order
//
// Example: (2 * (3 + 4) gets corrected to ((2 * 3) + 4)
// Example: ((1 + 2) * 3) gets corrected to (1 + (2 * 3)
func sortInfix(outer *OperatorNode) *OperatorNode {

	// outer: OP(OP(1 + 2) * 3)
	// outerOp: *
	// leftOp: +
	// left: OP(1 + 2)
	// res: OP(1 + OP(2 * 3)

	if left, ok := outer.Left.(*OperatorNode); ok {
		if infixPrio(outer.Operator) > infixPrio(left.Operator) {
			res := &OperatorNode{
				Operator: left.Operator,
				Left:     left.Left,
				Right: &OperatorNode{
					Operator: outer.Operator,
					Left:     left.Right,
					Right:    outer.Right,
				},
			}
			return res
		}
	}

	return outer
}
