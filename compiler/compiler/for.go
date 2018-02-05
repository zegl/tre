package compiler

import "github.com/zegl/tre/compiler/parser"

func (c *compiler) compileForNode(v parser.ForNode) {
	// TODO: create a new context-block for code running inside the for loop
	c.compile([]parser.Node{
		v.BeforeLoop,
	})

	// Check condition block
	checkCondBlock := c.contextBlock.Parent.NewBlock(getBlockName() + "-cond")

	// Loop body block
	loopBodyBlock := c.contextBlock.Parent.NewBlock(getBlockName() + "-body")

	// After loop block
	afterLoopBlock := c.contextBlock.Parent.NewBlock(getBlockName() + "-after")

	// Jump from BeforeLoop to checkCond
	c.contextBlock.NewBr(checkCondBlock)

	// Compile condition block
	c.contextBlock = checkCondBlock
	cmp := c.compileCondition(v.Condition)
	c.contextBlock.NewCondBr(cmp, loopBodyBlock, afterLoopBlock)

	// Compiler loop body
	c.contextBlock = loopBodyBlock
	c.compile(v.Block)
	c.compile([]parser.Node{v.AfterIteration})
	c.contextBlock.NewBr(checkCondBlock)

	// Set context to the new block after the loop
	c.contextBlock = afterLoopBlock
}
