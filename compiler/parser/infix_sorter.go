package parser

/*
https://golang.org/ref/spec#Operator_precedence

Precedence    Operator
    5             *  /  %  <<  >>  &  &^
    4             +  -  |  ^
    3             ==  !=  <  <=  >  >=
    2             &&
    1             ||
*/
var infixPrioMap = map[Operator]int{
	OP_MUL:         5,
	OP_DIV:         5,
	OP_REMAINDER:   5,
	OP_LEFT_SHIFT:  5,
	OP_RIGHT_SHIFT: 5,
	OP_BIT_AND:     5,
	OP_BIT_CLEAR:   5,

	OP_ADD:     4,
	OP_SUB:     4,
	OP_BIT_OR:  4,
	OP_BIT_XOR: 4,

	OP_EQ:   3,
	OP_NEQ:  3,
	OP_LT:   3,
	OP_LTEQ: 3,
	OP_GT:   3,
	OP_GTEQ: 3,

	OP_LOGICAL_AND: 2,

	OP_LOGICAL_OR: 1,
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
