package compiler

import (
	"fmt"

	"github.com/zegl/tre/parser"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
	"github.com/llir/llvm/ir/constant"
	"log"
)

type compiler struct {
	module *ir.Module

	// functions provided by the OS, such as printf
	externalFuncs map[string]*ir.Function

	// functions provided by the language, such as println
	globalFuncs map[string]*ir.Function
}

var (
	i8  = types.I8
	i32 = types.I32
	i64 = types.I64
)

func Compile(root parser.BlockNode) string {
	c := &compiler{
		module:        ir.NewModule(),
		externalFuncs: make(map[string]*ir.Function),
		globalFuncs:   make(map[string]*ir.Function),
	}

	c.addExternal()
	c.addGlobal()

	// TODO: Set automatically
	c.module.TargetTriple = "x86_64-apple-macosx10.13.0"

	mainFunc := c.module.NewFunction("main", i32)
	block := mainFunc.NewBlock("entry")

	c.compile(mainFunc, block, root.Instructions)

	// Print IR
	return fmt.Sprintln(c.module)
}

func (c *compiler) addExternal() {
	printfFunc := c.module.NewFunction("printf", i32, ir.NewParam("", types.NewPointer(i8)))
	printfFunc.Sig.Variadic = true
	c.externalFuncs["printf"] = printfFunc
}

func (c *compiler) addGlobal() {
	// TODO: Remove println method
	percentD := c.module.NewGlobalDef(getNextStringName(), constantString("%d\n"))
	percentD.IsConst = true

	printlnFunc := c.module.NewFunction("println", types.Void, ir.NewParam("num", i64))
	printlnFunc.Sig.Variadic = true

	c.globalFuncs["println"] = printlnFunc
	block := printlnFunc.NewBlock("entry")

	block.NewCall(c.externalFuncs["printf"], stringToi8Ptr(block, percentD), printlnFunc.Sig.Params[0])
	block.NewRet(nil)

	// Expose printf
	c.globalFuncs["printf"] = c.externalFuncs["printf"]
}

func (c *compiler) compile(function *ir.Function, block *ir.BasicBlock, instructions []parser.Node) {
	for _, i := range instructions {
		switch v := i.(type) {
		case parser.CallNode:
			var args []value.Value

			for _, vv := range v.Arguments {
				args = append(args, c.compileValue(block, vv))
			}

			block.NewCall(c.funcByName(v.Function), args...)
			break

		case parser.ConditionNode:
			xPtr := block.NewAlloca(i64)
			block.NewStore(c.compileValue(block, v.Cond.Left), xPtr)

			yPtr := block.NewAlloca(i64)
			block.NewStore(c.compileValue(block, v.Cond.Right), yPtr)

			cond := block.NewICmp(getConditionLLVMpred(v.Cond.Operator), block.NewLoad(xPtr), block.NewLoad(yPtr))

			afterBlock := function.NewBlock(getBlockName())
			trueBlock := function.NewBlock(getBlockName())
			falseBlock := function.NewBlock(getBlockName())

			c.compile(function, trueBlock, v.True)
			c.compile(function, falseBlock, v.False)

			trueBlock.NewBr(afterBlock)
			falseBlock.NewBr(afterBlock)

			block.NewCondBr(cond, trueBlock, falseBlock)

			block = afterBlock
			break

		default:
			log.Panicf("Unkown op: %+v", v)
			break
		}
	}

	// Return 0
	block.NewRet(constant.NewInt(0, i32))
}

func (c *compiler) funcByName(name string) *ir.Function {
	if f, ok := c.globalFuncs[name]; ok {
		return f
	}

	panic("funcByName: no such func: " + name)
}

func (c *compiler) compileValue(block *ir.BasicBlock, node parser.Node) value.Value {
	switch v := node.(type) {

	case parser.ConstantNode:
		switch v.Type {
		case parser.NUMBER:
			return constant.NewInt(v.Value, i64)
			break

		case parser.STRING:
			res := c.module.NewGlobalDef(getNextStringName(), constantString(v.ValueStr))
			res.IsConst = true
			return stringToi8Ptr(block, res)
			break
		}
		break

	case parser.OperatorNode:
		lPtr := block.NewAlloca(i64)
		block.NewStore(c.compileValue(block, v.Left), lPtr)

		rPtr := block.NewAlloca(i64)
		block.NewStore(c.compileValue(block, v.Right), rPtr)

		switch v.Operator {
		case parser.OP_ADD:
			return block.NewAdd(block.NewLoad(lPtr), block.NewLoad(rPtr))
			break
		case parser.OP_SUB:
			return block.NewSub(block.NewLoad(lPtr), block.NewLoad(rPtr))
			break
		case parser.OP_MUL:
			return block.NewMul(block.NewLoad(lPtr), block.NewLoad(rPtr))
			break
		case parser.OP_DIV:
			return block.NewSDiv(block.NewLoad(lPtr), block.NewLoad(rPtr)) // SDiv == Signed Division
			break
		}
		break
	}

	panic("compileValue fail")
}
