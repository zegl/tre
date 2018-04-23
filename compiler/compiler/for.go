package compiler

import "github.com/zegl/tre/compiler/parser"

func (c *Compiler) compileForNode(v parser.ForNode) {
	// TODO: create a new context-block for code running inside the for loop
	c.compile([]parser.Node{
		v.BeforeLoop,
	})

	// Check condition block
	checkCondBlock := c.contextBlock.Parent.NewBlock(getBlockName() + "-cond")

	// Loop body block
	loopBodyBlock := c.contextBlock.Parent.NewBlock(getBlockName() + "-body")
	loopAfterBodyBlock := c.contextBlock.Parent.NewBlock(getBlockName() + "-after-body")

	// After loop block
	afterLoopBlock := c.contextBlock.Parent.NewBlock(getBlockName() + "-after")

	// Push the break and continue stacks
	c.contextLoopBreak = append(c.contextLoopBreak, afterLoopBlock)
	c.contextLoopContinue = append(c.contextLoopContinue, loopAfterBodyBlock)

	// Jump from BeforeLoop to checkCond
	c.contextBlock.NewBr(checkCondBlock)

	// Compile condition block
	c.contextBlock = checkCondBlock
	cmp := c.compileOperatorNode(v.Condition)
	c.contextBlock.NewCondBr(cmp.Value, loopBodyBlock, afterLoopBlock)

	// Compiler loop body
	c.contextBlock = loopBodyBlock
	c.compile(v.Block)
	c.contextBlock.NewBr(loopAfterBodyBlock) // Jump to after body

	// After body block
	c.contextBlock = loopAfterBodyBlock
	c.compile([]parser.Node{v.AfterIteration})
	c.contextBlock.NewBr(checkCondBlock)

	// Set context to the new block after the loop
	c.contextBlock = afterLoopBlock

	// Pop break and continue
	c.contextLoopBreak = c.contextLoopBreak[0 : len(c.contextLoopBreak)-1]
	c.contextLoopContinue = c.contextLoopContinue[0 : len(c.contextLoopContinue)-1]
}

func (c *Compiler) compileBreakNode(v parser.BreakNode) {
	c.contextBlock.NewBr(c.contextLoopBreak[len(c.contextLoopBreak)-1])
}

func (c *Compiler) compileContinueNode(v parser.ContinueNode) {
	c.contextBlock.NewBr(c.contextLoopContinue[len(c.contextLoopContinue)-1])
}
