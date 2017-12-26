package compiler

import (
	"fmt"

	"github.com/zegl/tre/parser"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
	"github.com/llir/llvm/ir/constant"
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

	c.compile(root)

	// Print IR
	return fmt.Sprintln(c.module)
}

func (c *compiler) addExternal() {
	printfFunc := c.module.NewFunction("printf", i32, ir.NewParam("", types.NewPointer(i8)))
	printfFunc.Sig.Variadic = true
	c.externalFuncs["printf"] = printfFunc
}

func (c *compiler) addGlobal() {
	percentD := c.module.NewGlobalDef(".str0", constantString("%d\n"))
	percentD.IsConst = true

	printlnFunc := c.module.NewFunction("println", types.Void, ir.NewParam("num", i64))
	printlnFunc.Sig.Variadic = true

	c.globalFuncs["println"] = printlnFunc
	block := printlnFunc.NewBlock("entry")

	block.NewCall(c.externalFuncs["printf"], stringToi8Ptr(block, percentD), printlnFunc.Sig.Params[0])
	block.NewRet(nil)
}

func (c *compiler) compile(root parser.BlockNode) {
	mainFunc := c.module.NewFunction("main", i32)
	entry := mainFunc.NewBlock("entry")

	for _, i := range root.Instructions {
		switch v := i.(type) {
		case parser.CallNode:

			var args []value.Value

			for _, vv := range v.Arguments {
				args = append(args, c.compileValue(entry, vv))
			}

			entry.NewCall(c.funcByName(v.Function), args...)
			break
		}
	}

	// Return 0
	entry.NewRet(constant.NewInt(0, i32))
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
		if v.Type == parser.NUMBER {
			return constant.NewInt(v.Value, i64)
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

func constantString(in string) *constant.Array {
	var constants []constant.Constant

	for _, char := range in {
		constants = append(constants, constant.NewInt(int64(char), i8))
	}

	// null
	constants = append(constants, constant.NewInt(0, i8))

	s := constant.NewArray(constants...)
	s.CharArray = true

	return s
}

func stringToi8Ptr(block *ir.BasicBlock, src value.Value) *ir.InstGetElementPtr {
	return block.NewGetElementPtr(src, constant.NewInt(0, i64), constant.NewInt(0, i64))
}
