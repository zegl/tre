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

	contextFunc           *ir.Function
	contextBlock          *ir.BasicBlock
	contextBlockVariables map[string]value.Value
	contextFuncRetBlock   *ir.BasicBlock
	contextFuncRetVal     *ir.InstAlloca
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

	c.compile(root.Instructions)

	// Print IR
	return fmt.Sprintln(c.module)
}

func (c *compiler) addExternal() {
	printfFunc := c.module.NewFunction("printf", i32, ir.NewParam("", types.NewPointer(i8)))
	printfFunc.Sig.Variadic = true
	c.externalFuncs["printf"] = printfFunc
}

func (c *compiler) addGlobal() {
	// TODO: Add builtin methods here.

	// Expose printf
	c.globalFuncs["printf"] = c.externalFuncs["printf"]
}

func (c *compiler) compile(instructions []parser.Node) {
	for _, i := range instructions {
		block := c.contextBlock
		function := c.contextFunc

		switch v := i.(type) {
		case parser.ConditionNode:
			xPtr := block.NewAlloca(i64)
			block.NewStore(c.compileValue(v.Cond.Left), xPtr)

			yPtr := block.NewAlloca(i64)
			block.NewStore(c.compileValue(v.Cond.Right), yPtr)

			cond := block.NewICmp(getConditionLLVMpred(v.Cond.Operator), block.NewLoad(xPtr), block.NewLoad(yPtr))

			afterBlock := function.NewBlock(getBlockName() + "-after")
			trueBlock := function.NewBlock(getBlockName() + "-true")
			falseBlock := function.NewBlock(getBlockName() + "-false")

			block.NewCondBr(cond, trueBlock, falseBlock)

			c.contextBlock = trueBlock
			c.compile(v.True)

			// Jump to after-block if no terminator has been set (such as a return statement)
			if trueBlock.Term == nil {
				trueBlock.NewBr(afterBlock)
			}

			c.contextBlock = falseBlock
			c.compile(v.False)

			// Jump to after-block if no terminator has been set (such as a return statement)
			if falseBlock.Term == nil {
				falseBlock.NewBr(afterBlock)
			}

			c.contextBlock = afterBlock
			break

		case parser.DefineFuncNode:
			params := make([]*types.Param, len(v.Arguments))
			for k, par := range v.Arguments {
				params[k] = ir.NewParam(par.Name, i64)
			}

			funcRetType := types.Type(types.Void)
			if len(v.ReturnValues) == 1 {
				funcRetType = convertTypes(v.ReturnValues[0].Type)
			}

			// Create a new function, and add it to the list of global functions
			fn := c.module.NewFunction(v.Name, funcRetType, params...)
			c.globalFuncs[v.Name] = fn

			entry := fn.NewBlock(getBlockName())

			// There can only be one ret statement per function
			if len(v.ReturnValues) == 1 {
				// Allocate variable to return, allocated in the entry block
				c.contextFuncRetVal = entry.NewAlloca(funcRetType)

				// The return block contains only load + return instruction
				c.contextFuncRetBlock = fn.NewBlock(getBlockName() + "-return")
				c.contextFuncRetBlock.NewRet(c.contextFuncRetBlock.NewLoad(c.contextFuncRetVal))
			} else {
				// Unset to make sure that they are not accidentally used
				c.contextFuncRetBlock = nil
				c.contextFuncRetVal = nil
			}

			c.contextFunc = fn
			c.contextBlock = entry
			c.contextBlockVariables = make(map[string]value.Value)

			c.compile(v.Body)

			// Return void if there is no return type explicitly set
			if len(v.ReturnValues) == 0 {
				entry.NewRet(nil)
			}
			break

		case parser.ReturnNode:
			// Set value and jump to return block
			block.NewStore(c.compileValue(v.Val), c.contextFuncRetVal)
			block.NewBr(c.contextFuncRetBlock)
			break

		case parser.AllocNode:
			allocVal := c.compileValue(v.Val)
			alloc := block.NewAlloca(allocVal.Type())
			alloc.SetName(v.Name)
			block.NewStore(allocVal, alloc)
			c.contextBlockVariables[v.Name] = alloc
			break

		default:
			c.compileValue(v)
			break
		}
	}
}

func (c *compiler) funcByName(name string) *ir.Function {
	if f, ok := c.globalFuncs[name]; ok {
		return f
	}

	panic("funcByName: no such func: " + name)
}

func (c *compiler) compileValue(node parser.Node) value.Value {
	block := c.contextBlock

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
		block.NewStore(c.compileValue(v.Left), lPtr)

		rPtr := block.NewAlloca(i64)
		block.NewStore(c.compileValue(v.Right), rPtr)

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

	case parser.NameNode:
		// Any parameter?
		for _, param := range block.Parent.Params() {
			if param.Name == v.Name {
				return param
			}
		}

		// Named variable in this block?
		if val, ok := c.contextBlockVariables[v.Name]; ok {
			return block.NewLoad(val)
		}

		panic("undefined variable: " + v.Name)
		break

	case parser.CallNode:
		var args []value.Value

		for _, vv := range v.Arguments {
			args = append(args, c.compileValue(vv))
		}

		return block.NewCall(c.funcByName(v.Function), args...)
		break
	}

	panic("compileValue fail: " + fmt.Sprintf("%+v", node))
}
