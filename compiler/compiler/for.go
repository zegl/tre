package compiler

import (
	"fmt"
	"github.com/zegl/tre/compiler/compiler/name"
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
	checkCondBlock := c.contextBlock.Parent.NewBlock(name.Block() + "-cond")

	// Loop body block
	loopBodyBlock := c.contextBlock.Parent.NewBlock(name.Block() + "-body")
	loopAfterBodyBlock := c.contextBlock.Parent.NewBlock(name.Block() + "-after-body")

	// After loop block
	afterLoopBlock := c.contextBlock.Parent.NewBlock(name.Block() + "-after")

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
	var modifiedBlock []parser.Node

	// The node/value that we're iterating over
	var rangeItem parser.Node

	if forAlloc, ok := v.BeforeLoop.(*parser.AllocNode); ok {
		forAllocRange := forAlloc.Val.(*parser.RangeNode)
		rangeItem = forAllocRange.Item
	} else if forRange, ok := v.BeforeLoop.(*parser.RangeNode); ok {
		rangeItem = forRange.Item
	} else {
		panic("unexpected for/range beforeLoop type: " + fmt.Sprintf("%T %+v", v.BeforeLoop, v.BeforeLoop))
	}

	// Alloc the value of rangeItem and save it in a variable
	rangeItemName := name.Var("range-item")
	c.compileAllocNode(&parser.AllocNode{
		Name: rangeItemName,
		Val:  rangeItem,
	})

	// Call and alloc len() and save it in a variable
	forItemLenName := name.Var("range-item-len")
	c.compileAllocNode(&parser.AllocNode{
		Name: forItemLenName,
		Val: &parser.CallNode{
			Function:  &parser.NameNode{Name: "len"},
			Arguments: []parser.Node{&parser.NameNode{Name: rangeItemName}},
		},
	})

	forKeyName := name.Var("for-key")

	// Ranges that use the key and value
	if forAlloc, ok := v.BeforeLoop.(*parser.AllocNode); ok {
		var keyName string
		if forAlloc.MultiNames == nil {
			keyName = forAlloc.Name
		} else {
			keyName = forAlloc.MultiNames.Names[0].Name
		}

		// Assignment of key
		modifiedBlock = append(modifiedBlock, &parser.AllocNode{
			Name: keyName,
			Val:  &parser.NameNode{Name: forKeyName},
		})

		// Assignment of value
		if forAlloc.MultiNames != nil && len(forAlloc.MultiNames.Names) >= 2 {
			modifiedBlock = append(modifiedBlock, &parser.AllocNode{
				Name: forAlloc.MultiNames.Names[1].Name,
				Val:  &parser.LoadArrayElement{Array: &parser.NameNode{Name: rangeItemName}, Pos: &parser.NameNode{Name: forKeyName}},
			})
		}
	}
	modifiedBlock = append(modifiedBlock, v.Block...)

	typeCastedKey := &parser.TypeCastNode{
		Type: &parser.SingleTypeNode{SourceName: "int32", TypeName: "int32"},
		Val:  &parser.NameNode{Name: forKeyName},
	}

	c.compileForThreeType(&parser.ForNode{
		BeforeLoop: &parser.AllocNode{Name: forKeyName, Val: &parser.ConstantNode{Type: parser.NUMBER, Value: 0}},

		Condition: &parser.OperatorNode{
			Left:     typeCastedKey,
			Operator: parser.OP_LT,
			Right: &parser.NameNode{
				Name: forItemLenName,
			},
		},

		AfterIteration: &parser.AssignNode{
			Target: []parser.Node{&parser.NameNode{Name: forKeyName}},
			Val: []parser.Node{&parser.OperatorNode{
				Left:     &parser.NameNode{Name: forKeyName},
				Operator: parser.OP_ADD,
				Right:    &parser.ConstantNode{Type: parser.NUMBER, Value: 1},
			}},
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
