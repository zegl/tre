package compiler

import (
	"github.com/zegl/tre/compiler/parser"
)

func (c *Compiler) compileForNode(v *parser.ForNode) {
	if v.IsThreeTypeFor {
		c.compileForThreeType(v)
		return
	}

	c.compileForRange(v)
}

func (c *Compiler) compileForThreeType(v *parser.ForNode) {
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

func (c *Compiler) compileForRange(v *parser.ForNode) {
	// A for range that iterates over a slice is just syntactic sugar
	// for k, v := range a
	// for k := 0; k < len(a); k++ { v := a[k] }


	forAlloc := v.BeforeLoop.(*parser.AllocNode)

	forAllocRange := forAlloc.Val.(*parser.RangeNode)

	typeCastedKey := &parser.TypeCastNode{
		Type: &parser.SingleTypeNode{SourceName: "int32", TypeName: "int32"},
		Val: &parser.NameNode{Name: "for-key"},
	}

	modifiedBlock := []parser.Node{
		&parser.AllocNode{Name: forAlloc.MultiNames.Names[0].Name, Val: &parser.NameNode{Name: "for-key"}},
		&parser.AllocNode{Name: forAlloc.MultiNames.Names[1].Name, Val:
			&parser.LoadArrayElement{Array: forAllocRange.Item, Pos: &parser.NameNode{Name: "for-key"}},
		},
	}

	modifiedBlock = append(modifiedBlock, v.Block...)

	c.compileForThreeType(&parser.ForNode{
		BeforeLoop: &parser.AllocNode{Name: "for-key", Val: &parser.ConstantNode{Type: parser.NUMBER, Value: 0}},

		Condition: &parser.OperatorNode{
			Left: typeCastedKey,
			Operator: parser.OP_LT,
			Right: &parser.CallNode{
					Function: &parser.NameNode{Name: "len"},
					Arguments: []parser.Node{forAllocRange.Item},
			},
		},

		AfterIteration: &parser.AssignNode{
			Target: &parser.NameNode{Name: "for-key"},
			Val: &parser.OperatorNode{
				Left: &parser.NameNode{Name: "for-key"},
				Operator: parser.OP_ADD,
				Right: &parser.ConstantNode{Type: parser.NUMBER, Value: 1},
			},
		},

		Block: modifiedBlock,
	})
}

func (c *Compiler) compileBreakNode(v *parser.BreakNode) {
	c.contextBlock.NewBr(c.contextLoopBreak[len(c.contextLoopBreak)-1])
}

func (c *Compiler) compileContinueNode(v *parser.ContinueNode) {
	c.contextBlock.NewBr(c.contextLoopContinue[len(c.contextLoopContinue)-1])
}
