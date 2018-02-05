package parser

import (
	"fmt"

	"github.com/zegl/tre/compiler/lexer"
)

func (p *parser) parseFor() ForNode {
	res := ForNode{}

	p.i++
	beforeLoop := p.parseUntil(lexer.Item{Type: lexer.SEPARATOR, Val: ";"})
	if len(beforeLoop) != 1 {
		panic("Expected only one beforeLoop in for loop")
	}
	res.BeforeLoop = beforeLoop[0]

	p.i++
	loopCondition := p.parseUntil(lexer.Item{Type: lexer.SEPARATOR, Val: ";"})
	if len(loopCondition) != 1 {
		panic("Expected only one condition in for loop")
	}

	if conditionNode, ok := loopCondition[0].(OperatorNode); ok {
		res.Condition = conditionNode
	} else {
		panic(fmt.Sprintf("Expected OperatorNode in for loop. Got: %T: %+v", loopCondition[0], loopCondition[0]))
	}

	p.i++
	afterIteration := p.parseUntil(lexer.Item{Type: lexer.SEPARATOR, Val: "{"})
	if len(afterIteration) != 1 {
		panic("Expected only one afterIteration in for loop")
	}
	res.AfterIteration = afterIteration[0]

	p.i++
	res.Block = p.parseUntil(lexer.Item{Type: lexer.SEPARATOR, Val: "}"})

	return res
}
