package parser

import (
	"fmt"
	"github.com/zegl/tre/compiler/lexer"
)

// ForNode creates a new for-loop
type ForNode struct {
	baseNode

	BeforeLoop     Node
	Condition      *OperatorNode
	AfterIteration Node
	Block          []Node

	IsThreeTypeFor bool
}

func (f ForNode) String() string {
	return fmt.Sprintf("For(%s; %s; %s) {\n\t%s\n}", f.BeforeLoop, f.Condition, f.AfterIteration, f.Block)
}

func (p *parser) parseFor() *ForNode {
	res := &ForNode{}

	p.i++
	beforeLoop, reachedItem := p.parseUntilEither([]lexer.Item{
		{Type: lexer.OPERATOR, Val: ";"}, // three type for
		{Type: lexer.OPERATOR, Val: "{"}, // range type for
	})

	if len(beforeLoop) != 1 {
		panic("Expected only one beforeLoop in for loop")
	}

	isThreeTypeFor := false
	if reachedItem.Val == ";" {
		isThreeTypeFor = true
		res.IsThreeTypeFor = true
	}

	res.BeforeLoop = beforeLoop[0]

	if isThreeTypeFor {
		p.i++
		loopCondition := p.parseUntil(lexer.Item{Type: lexer.OPERATOR, Val: ";"})
		if len(loopCondition) != 1 {
			panic("Expected only one condition in for loop")
		}

		if conditionNode, ok := loopCondition[0].(*OperatorNode); ok {
			res.Condition = conditionNode
		} else {
			panic(fmt.Sprintf("Expected OperatorNode in for loop. Got: %T: %+v", loopCondition[0], loopCondition[0]))
		}

		p.i++
		afterIteration := p.parseUntil(lexer.Item{Type: lexer.OPERATOR, Val: "{"})
		if len(afterIteration) != 1 {
			panic("Expected only one afterIteration in for loop")
		}
		res.AfterIteration = afterIteration[0]
	}

	p.i++
	res.Block = p.parseUntil(lexer.Item{Type: lexer.OPERATOR, Val: "}"})

	return res
}
