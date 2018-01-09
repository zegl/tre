package compiler

import (
	"fmt"

	"github.com/zegl/tre/parser"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
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

func ptrTypeType(val value.Value) types.Type {
	if t, valIsPtr := val.Type().(*types.PointerType); valIsPtr {
		return t.Elem
	}
	panic("ptrTypeType is not pointer type")
}

func (c *compiler) compile(instructions []parser.Node) {
	for _, i := range instructions {
		block := c.contextBlock
		function := c.contextFunc

		switch v := i.(type) {
		case parser.ConditionNode:
			xVal := c.compileValue(v.Cond.Left)
			xPtr := block.NewAlloca(ptrTypeType(xVal))
			block.NewStore(block.NewLoad(xVal), xPtr)

			yVal := c.compileValue(v.Cond.Right)
			yPtr := block.NewAlloca(ptrTypeType(yVal))
			block.NewStore(block.NewLoad(yVal), yPtr)

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
				params[k] = ir.NewParam(par.Name+"-parameter", typeStringToLLVM(par.Type))
			}

			funcRetType := types.Type(types.Void)
			if len(v.ReturnValues) == 1 {
				funcRetType = typeStringToLLVM(v.ReturnValues[0].Type)
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

			// Allocate all parameters
			for i, param := range params {
				paramName := v.Arguments[i].Name
				paramPtr := entry.NewAlloca(typeStringToLLVM(v.Arguments[i].Type))
				paramPtr.SetName(paramName)
				entry.NewStore(param, paramPtr)
				c.contextBlockVariables[paramName] = paramPtr
			}

			c.compile(v.Body)

			// Return void if there is no return type explicitly set
			if len(v.ReturnValues) == 0 {
				entry.NewRet(nil)
			}
			break

		case parser.ReturnNode:
			// Set value and jump to return block
			block.NewStore(block.NewLoad(c.compileValue(v.Val)), c.contextFuncRetVal)
			block.NewBr(c.contextFuncRetBlock)
			break

		case parser.AllocNode:
			val := c.compileValue(v.Val)
			allocVal := block.NewLoad(val)
			alloc := block.NewAlloca(allocVal.Type())
			alloc.SetName(v.Name)
			block.NewStore(allocVal, alloc)
			c.contextBlockVariables[v.Name] = alloc
			break

		case parser.AssignNode:
			allocVal := block.NewLoad(c.compileValue(v.Val))
			dst := c.varByName(v.Name)
			block.NewStore(allocVal, dst)
			break

		case parser.DefineTypeNode:

			if structNode, ok := v.Type.(*parser.StructTypeNode); ok {

				var structTypes []types.Type
				for _, t := range structNode.Types {
					if singleTypeNode, ok := t.(*parser.SingleTypeNode); ok {
						structTypes = append(structTypes, convertTypes(singleTypeNode.TypeName))
					} else {
						panic("unable to define node Type. nested structs are not supported")
					}
				}

				typeConvertMap[v.Name] = types.NewStruct(structTypes...)
				// constant.NewStruct()
				break
			}

			panic("unable to define node Type")
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

func (c *compiler) varByName(name string) value.Value {
	// Named variable in this block?
	if val, ok := c.contextBlockVariables[name]; ok {
		return val
	}

	panic("undefined variable: " + name)
}

func (c *compiler) compileValue(node parser.Node) value.Value {
	block := c.contextBlock

	switch v := node.(type) {

	case parser.ConstantNode:
		switch v.Type {
		case parser.NUMBER:
			dst := block.NewAlloca(i64)
			num := constant.NewInt(v.Value, i64)
			block.NewStore(num, dst)
			return dst
			break

		case parser.STRING:
			res := c.module.NewGlobalDef(getNextStringName(), constantString(v.ValueStr))
			res.IsConst = true
			return stringToi8Ptr(block, res)
			break
		}
		break

	case parser.OperatorNode:
		left := c.compileValue(v.Left)
		right := c.compileValue(v.Right)

		left = block.NewLoad(left)
		right = block.NewLoad(right)

		res := block.NewAlloca(i64)

		switch v.Operator {
		case parser.OP_ADD:
			block.NewStore(block.NewAdd(left, right), res)
			return res
			break
		case parser.OP_SUB:
			block.NewStore(block.NewSub(left, right), res)
			return res
			break
		case parser.OP_MUL:
			block.NewStore(block.NewMul(left, right), res)
			return res
			break
		case parser.OP_DIV:
			block.NewStore(block.NewSDiv(left, right), res) // SDiv == Signed Division
			return res
			break
		}
		break

	case parser.NameNode:
		return c.varByName(v.Name)
		break

	case parser.CallNode:
		var args []value.Value

		for _, vv := range v.Arguments {
			val := c.compileValue(vv)
			_, valIsPtr := val.Type().(*types.PointerType)

			if val.Type().Equal(types.NewPointer(types.I8)) {
				// Strings
				args = append(args, val)
			} else if valIsPtr {
				// Dereference value
				args = append(args, block.NewLoad(val))
			} else {
				// Use as is
				args = append(args, val)
			}
		}

		fn := c.funcByName(v.Function)

		// Create variable to save result in
		retVal := block.NewAlloca(fn.Sig.Ret)

		// Call function and store results
		block.NewStore(block.NewCall(fn, args...), retVal)
		return retVal
		break

	case parser.TypeCastNode:
		val := c.compileValue(v.Val)

		currentType := ptrTypeType(val)

		current, ok := currentType.(*types.IntType)
		if !ok {
			panic("TypeCast origin must be int type")
		}

		target, ok := typeStringToLLVM(v.Type).(*types.IntType)
		if !ok {
			panic("TypeCast target must be int type")
		}

		// Same size, nothing to do here
		if current.Size == target.Size {
			return val
		}

		res := block.NewAlloca(target)

		if current.Size < target.Size {
			block.NewStore(block.NewSExt(block.NewLoad(val), target), res)
		} else {
			block.NewStore(block.NewTrunc(block.NewLoad(val), target), res)
		}

		return res
		break
	}

	panic("compileValue fail: " + fmt.Sprintf("%+v", node))
}
