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

	afterSwitch := c.contextBlock.Parent.NewBlock(name.Block() + "after-switch")
	defaultCase := c.contextBlock.Parent.NewBlock(name.Block() + "switch-default")
	defaultCase.NewBr(afterSwitch)

	for _, parseCase := range v.Cases {
		item := c.compileValue(parseCase.Condition)
		preCaseBlock := c.contextBlock
		caseBlock := c.contextBlock.Parent.NewBlock(name.Block() + "case")
		c.contextBlock = caseBlock
		c.compile(parseCase.Body)
		caseBlock.Term = ir.NewBr(afterSwitch)
		c.contextBlock = preCaseBlock
		cases = append(cases, ir.NewCase(item.Value.(constant.Constant), caseBlock))
	}

	val := switchItem.Value
	if switchItem.IsVariable {
		val = c.contextBlock.NewLoad(val)
	}

	c.contextBlock.Term = c.contextBlock.NewSwitch(val, defaultCase, cases...)

	c.contextBlock = afterSwitch
}
