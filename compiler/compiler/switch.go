package compiler

import (
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"

	"github.com/zegl/tre/compiler/compiler/internal"
	"github.com/zegl/tre/compiler/compiler/name"
	"github.com/zegl/tre/compiler/parser"
)

func (c *Compiler) compileSwitchNode(v *parser.SwitchNode) {
	switchItem := c.compileValue(v.Item)

	var cases []*ir.Case
	caseBlocks := make([]*ir.Block, len(v.Cases))

	afterSwitch := c.contextBlock.Parent.NewBlock(name.Block() + "after-switch")

	// break inside of a switch jumps to after the switch
	c.contextLoopBreak = append(c.contextLoopBreak, afterSwitch)

	// build default case
	defaultCase := c.contextBlock.Parent.NewBlock(name.Block() + "switch-default")
	preDefaultBlock := c.contextBlock
	c.contextBlock = defaultCase
	if v.DefaultBody != nil {
		c.compile(v.DefaultBody)
	}
	if c.contextBlock.Term == nil {
		c.contextBlock.NewBr(afterSwitch)
	}
	c.contextBlock = preDefaultBlock

	// Parse all cases
	caseEndBlocks := make([]*ir.Block, len(v.Cases))
	for caseIndex, parseCase := range v.Cases {
		preCaseBlock := c.contextBlock
		caseBlock := c.contextBlock.Parent.NewBlock(name.Block() + "case")
		c.contextBlock = caseBlock
		c.compile(parseCase.Body)
		caseEndBlocks[caseIndex] = c.contextBlock
		c.contextBlock = preCaseBlock

		caseBlocks[caseIndex] = caseBlock

		for _, cond := range parseCase.Conditions {
			item := c.compileValue(cond)
			cases = append(cases, ir.NewCase(item.Value.(constant.Constant), caseBlock))
		}
	}

	for caseIndex, parseCase := range v.Cases {
		endBlock := caseEndBlocks[caseIndex]
		if endBlock.Term != nil {
			continue
		}
		if parseCase.Fallthrough {
			// Jump to the next case body
			endBlock.NewBr(caseBlocks[caseIndex+1])
		} else {
			// Jump to after switch
			endBlock.NewBr(afterSwitch)
		}
	}

	c.contextLoopBreak = c.contextLoopBreak[:len(c.contextLoopBreak)-1]

	val := internal.LoadIfVariable(c.contextBlock, switchItem)
	c.contextBlock.Term = c.contextBlock.NewSwitch(val, defaultCase, cases...)

	c.contextBlock = afterSwitch
}
