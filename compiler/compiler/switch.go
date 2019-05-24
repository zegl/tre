package compiler

import (
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"

	"github.com/zegl/tre/compiler/compiler/name"
	"github.com/zegl/tre/compiler/parser"
)

func (c *Compiler) compileSwitchNode(v *parser.SwitchNode) {
	switchItem := c.compileValue(v.Item)

	var cases []*ir.Case
	caseBlocks := make([]*ir.Block, len(v.Cases))

	afterSwitch := c.contextBlock.Parent.NewBlock(name.Block() + "after-switch")

	// build default case
	defaultCase := c.contextBlock.Parent.NewBlock(name.Block() + "switch-default")
	if v.DefaultBody != nil {
		preDefaultBlock := c.contextBlock
		c.contextBlock = defaultCase
		c.compile(v.DefaultBody)
		c.contextBlock = preDefaultBlock
	}
	defaultCase.NewBr(afterSwitch)

	// Parse all cases
	for caseIndex, parseCase := range v.Cases {
		preCaseBlock := c.contextBlock
		caseBlock := c.contextBlock.Parent.NewBlock(name.Block() + "case")
		c.contextBlock = caseBlock
		c.compile(parseCase.Body)
		c.contextBlock = preCaseBlock

		caseBlocks[caseIndex] = caseBlock

		for _, cond := range parseCase.Conditions {
			item := c.compileValue(cond)
			cases = append(cases, ir.NewCase(item.Value.(constant.Constant), caseBlock))
		}
	}

	for caseIndex, parseCase := range v.Cases {
		if parseCase.Fallthrough {
			// Jump to the next case body
			caseBlocks[caseIndex].Term = ir.NewBr(caseBlocks[caseIndex+1])
		} else {
			// Jump to after switch
			caseBlocks[caseIndex].Term = ir.NewBr(afterSwitch)
		}
	}

	val := switchItem.Value
	if switchItem.IsVariable {
		val = c.contextBlock.NewLoad(val)
	}

	c.contextBlock.Term = c.contextBlock.NewSwitch(val, defaultCase, cases...)

	c.contextBlock = afterSwitch
}
