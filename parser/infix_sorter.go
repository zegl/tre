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
func sortInfix(left, right OperatorNode) OperatorNode {
	if infixPrio(left.Operator) > infixPrio(right.Operator) {
		return OperatorNode{
			Operator: right.Operator,
			Left: OperatorNode{
				Operator: left.Operator,
				Left:     left.Left,
				Right:    right.Left,
			},
			Right: right.Right,
		}
	}

	return left
}
