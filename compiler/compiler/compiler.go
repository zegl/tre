package compiler

import (
	"fmt"
	"os"
	"runtime"

	"github.com/zegl/tre/compiler/compiler/internal"
	"github.com/zegl/tre/compiler/compiler/strings"
	"github.com/zegl/tre/compiler/compiler/types"
	"github.com/zegl/tre/compiler/compiler/value"
	"github.com/zegl/tre/compiler/parser"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	llvmTypes "github.com/llir/llvm/ir/types"
	llvmValue "github.com/llir/llvm/ir/value"
)

type compiler struct {
	module *ir.Module

	// functions provided by the OS, such as printf
	externalFuncs map[string]*ir.Function

	// functions provided by the language, such as println
	globalFuncs map[string]*types.Function

	contextFunc           *ir.Function
	contextBlock          *ir.BasicBlock
	contextBlockVariables map[string]value.Value

	// What a break or continue should resolve to
	contextLoopBreak    []*ir.BasicBlock
	contextLoopContinue []*ir.BasicBlock
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
		globalFuncs:   make(map[string]*types.Function),

		contextLoopBreak:    make([]*ir.BasicBlock, 0),
		contextLoopContinue: make([]*ir.BasicBlock, 0),
	}

	c.addExternal()
	c.addGlobal()

	// Triple examples:
	// x86_64-apple-macosx10.13.0
	// x86_64-pc-linux-gnu
	var targetTriple [2]string

	switch runtime.GOARCH {
	case "amd64":
		targetTriple[0] = "x86_64"
	default:
		panic("unsupported GOARCH: " + runtime.GOARCH)
	}

	switch runtime.GOOS {
	case "darwin":
		targetTriple[1] = "apple-macosx10.13.0"
	case "linux":
		targetTriple[1] = "pc-linux-gnu"
	default:
		panic("unsupported GOOS: " + runtime.GOOS)
	}

	c.module.TargetTriple = fmt.Sprintf("%s-%s", targetTriple[0], targetTriple[1])

	c.compile(root.Instructions)

	// Print IR
	return fmt.Sprintln(c.module)
}

func (c *compiler) addExternal() {
	printfFunc := c.module.NewFunction("printf", i32.LLVM(), ir.NewParam("", llvmTypes.NewPointer(i8.LLVM())))
	printfFunc.Sig.Variadic = true
	c.externalFuncs["printf"] = printfFunc

	c.externalFuncs["strcat"] = c.module.NewFunction("strcat",
		llvmTypes.NewPointer(i8.LLVM()),
		ir.NewParam("", llvmTypes.NewPointer(i8.LLVM())),
		ir.NewParam("", llvmTypes.NewPointer(i8.LLVM())))

	c.externalFuncs["strcpy"] = c.module.NewFunction("strcpy",
		llvmTypes.NewPointer(i8.LLVM()),
		ir.NewParam("", llvmTypes.NewPointer(i8.LLVM())),
		ir.NewParam("", llvmTypes.NewPointer(i8.LLVM())))

	c.externalFuncs["strncpy"] = c.module.NewFunction("strncpy",
		llvmTypes.NewPointer(i8.LLVM()),
		ir.NewParam("", llvmTypes.NewPointer(i8.LLVM())),
		ir.NewParam("", llvmTypes.NewPointer(i8.LLVM())),
		ir.NewParam("", i64.LLVM()),
	)

	c.externalFuncs["strndup"] = c.module.NewFunction("strndup",
		llvmTypes.NewPointer(i8.LLVM()),
		ir.NewParam("", llvmTypes.NewPointer(i8.LLVM())),
		ir.NewParam("", i64.LLVM()),
	)

	c.externalFuncs["exit"] = c.module.NewFunction("exit",
		llvmTypes.Void,
		ir.NewParam("", i32.LLVM()),
	)
}

func (c *compiler) addGlobal() {
	types.ModuleStringType = c.module.NewType("string", internal.String())

	// printf
	c.globalFuncs["printf"] = &types.Function{
		LlvmFunction: c.externalFuncs["printf"],
		FunctionName: "printf",
	}

	// println
	c.globalFuncs["println"] = &types.Function{
		LlvmFunction: internal.Println(types.ModuleStringType, c.externalFuncs["printf"], c.module),
		FunctionName: "println",
	}
	c.module.AppendFunction(c.globalFuncs["println"].LlvmFunction)

	// len_string
	strLen := internal.StringLen(types.ModuleStringType)
	c.globalFuncs["len_string"] = &types.Function{
		LlvmFunction: strLen,
		ReturnType:   types.I64,
	}
	c.module.AppendFunction(strLen)
}

func (c *compiler) compile(instructions []parser.Node) {
	for _, i := range instructions {
		block := c.contextBlock

		switch v := i.(type) {
		case parser.ConditionNode:
			cond := c.compileCondition(v.Cond)

			afterBlock := c.contextFunc.NewBlock(getBlockName() + "-after")
			trueBlock := c.contextFunc.NewBlock(getBlockName() + "-true")
			falseBlock := afterBlock

			if len(v.False) > 0 {
				falseBlock = c.contextFunc.NewBlock(getBlockName() + "-false")
			}

			block.NewCondBr(cond.Value, trueBlock, falseBlock)

			c.contextBlock = trueBlock
			c.compile(v.True)

			// Jump to after-block if no terminator has been set (such as a return statement)
			if trueBlock.Term == nil {
				trueBlock.NewBr(afterBlock)
			}

			if len(v.False) > 0 {
				c.contextBlock = falseBlock
				c.compile(v.False)

				// Jump to after-block if no terminator has been set (such as a return statement)
				if falseBlock.Term == nil {
					falseBlock.NewBr(afterBlock)
				}
			}

			c.contextBlock = afterBlock
			break

		case parser.DefineFuncNode:

			var compiledName string

			if v.IsMethod {
				// Add the type that we're a method on as the first argument
				v.Arguments = append(v.Arguments, parser.NameNode{
					Name: v.InstanceName,
					Type: v.MethodOnType,
				})

				// Change the name of our function
				compiledName = "method_" + v.MethodOnType.TypeName + "_" + v.Name
			} else {
				compiledName = v.Name
			}

			params := make([]*llvmTypes.Param, len(v.Arguments))
			for k, par := range v.Arguments {
				params[k] = ir.NewParam(par.Name, parserTypeToType(par.Type).LLVM())
			}

			var funcRetType types.Type = types.Void

			if len(v.ReturnValues) == 1 {
				funcRetType = parserTypeToType(v.ReturnValues[0].Type)
			}

			// TODO: Only do this in the main package
			if v.Name == "main" {
				if len(v.ReturnValues) != 0 {
					panic("main func can not have a return type")
				}

				funcRetType = types.I32
			}

			// Create a new function, and add it to the list of global functions
			fn := c.module.NewFunction(compiledName, funcRetType.LLVM(), params...)

			typesFunc := &types.Function{
				LlvmFunction: fn,
				ReturnType:   funcRetType,
				FunctionName: v.Name,
			}

			// Save as a method on the type
			if v.IsMethod {
				if t, ok := typeConvertMap[v.MethodOnType.TypeName]; ok {
					t.AddMethod(v.Name, &types.Method{
						Function:        typesFunc,
						PointerReceiver: false,
						MethodName:      v.Name,
					})
				} else {
					panic("save method on type failed")
				}

			} else {
				c.globalFuncs[v.Name] = typesFunc
			}

			entry := fn.NewBlock(getBlockName())

			c.contextFunc = fn
			c.contextBlock = entry
			c.contextBlockVariables = make(map[string]value.Value)

			// Save all parameters in the block mapping
			for i, param := range params {
				paramName := v.Arguments[i].Name
				dataType := parserTypeToType(v.Arguments[i].Type)

				// Structs needs to be pointer-allocated
				if _, ok := param.Type().(*llvmTypes.StructType); ok {
					paramPtr := entry.NewAlloca(dataType.LLVM())
					entry.NewStore(param, paramPtr)

					c.contextBlockVariables[paramName] = value.Value{
						Value:        paramPtr,
						Type:         dataType,
						PointerLevel: 1,
					}

					continue
				}

				// TODO: Using 0 as the pointer level here might not be correct
				c.contextBlockVariables[paramName] = value.Value{
					Value:        param,
					Type:         dataType,
					PointerLevel: 0,
				}
			}

			c.compile(v.Body)

			// Return void if there is no return type explicitly set
			if len(v.ReturnValues) == 0 {
				c.contextBlock.NewRet(nil)
			}

			// Return 0 by default in main func
			if v.Name == "main" {
				c.contextBlock.NewRet(constant.NewInt(0, types.I32.LLVM()))
			}

			break

		case parser.ReturnNode:
			// Set value and jump to return block
			val := c.compileValue(v.Val)

			if val.PointerLevel > 0 {
				block.NewRet(block.NewLoad(val.Value))
				break
			}

			block.NewRet(val.Value)
			break

		case parser.AllocNode:

			// Allocate from type
			if typeNode, ok := v.Val.(parser.TypeNode); ok {
				treType := parserTypeToType(typeNode)

				alloc := block.NewAlloca(treType.LLVM())
				alloc.SetName(v.Name)

				c.contextBlockVariables[v.Name] = value.Value{
					Value:        alloc,
					Type:         treType,
					PointerLevel: 1, // This is probably not always correct
				}
				break
			}

			// Allocate from value
			val := c.compileValue(v.Val)
			llvmVal := val.Value

			if val.PointerLevel > 0 {
				llvmVal = block.NewLoad(llvmVal)
			}

			alloc := block.NewAlloca(llvmVal.Type())
			alloc.SetName(v.Name)
			block.NewStore(llvmVal, alloc)

			c.contextBlockVariables[v.Name] = value.Value{
				Type:         val.Type,
				Value:        alloc,
				PointerLevel: 1, // TODO
			}
			break

		case parser.AssignNode:
			var dst value.Value

			// TODO: Remove AssignNode.Name
			if len(v.Name) > 0 {
				dst = c.varByName(v.Name)
			} else {
				dst = c.compileValue(v.Target)
			}

			llvmDst := dst.Value

			// Allocate from type
			if typeNode, ok := v.Val.(parser.TypeNode); ok {
				if singleTypeNode, ok := typeNode.(parser.SingleTypeNode); ok {
					alloc := block.NewAlloca(parserTypeToType(singleTypeNode).LLVM())
					block.NewStore(block.NewLoad(alloc), llvmDst)
					break
				}

				panic("AssignNode from non TypeNode is not allowed")
			}

			// Allocate from value
			val := c.compileValue(v.Val)
			llvmV := val.Value

			if val.PointerLevel > 0 {
				llvmV = block.NewLoad(llvmV)
			}

			block.NewStore(llvmV, llvmDst)
			break

		case parser.DefineTypeNode:
			t := parserTypeToType(v.Type)

			// Add type to module and override the structtype to use the named
			// type in the module
			if structType, ok := t.(*types.Struct); ok {
				structType.Type = c.module.NewType(v.Name, t.LLVM())
			}

			// Add to tre mapping
			typeConvertMap[v.Name] = t

		case parser.DeclarePackageNode:
			// TODO: Make use of it
			break

		case parser.ForNode:
			c.compileForNode(v)
			break

		case parser.BreakNode:
			c.compileBreakNode(v)
			break

		case parser.ContinueNode:
			c.compileContinueNode(v)
			break

		default:
			c.compileValue(v)
			break
		}
	}
}

func (c *compiler) funcByName(name string) *types.Function {
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
	switch v := node.(type) {

	case parser.ConstantNode:
		switch v.Type {
		case parser.NUMBER:
			return value.Value{
				Value:        constant.NewInt(v.Value, i64.LLVM()),
				Type:         i64,
				PointerLevel: 0,
			}

		case parser.STRING:
			constString := c.module.NewGlobalDef(strings.NextStringName(), strings.Constant(v.ValueStr))
			constString.IsConst = true

			alloc := c.contextBlock.NewAlloca(typeConvertMap["string"].LLVM())

			// Save length of the string
			lenItem := c.contextBlock.NewGetElementPtr(alloc, constant.NewInt(0, i32.LLVM()), constant.NewInt(0, i32.LLVM()))
			c.contextBlock.NewStore(constant.NewInt(int64(len(v.ValueStr)), i64.LLVM()), lenItem)

			// Save i8* version of string
			strItem := c.contextBlock.NewGetElementPtr(alloc, constant.NewInt(0, i32.LLVM()), constant.NewInt(1, i32.LLVM()))
			c.contextBlock.NewStore(strings.Toi8Ptr(c.contextBlock, constString), strItem)

			return value.Value{
				Value:        c.contextBlock.NewLoad(alloc),
				Type:         types.String,
				PointerLevel: 0,
			}
		}
		break

	case parser.OperatorNode:
		left := c.compileValue(v.Left)
		leftLLVM := left.Value

		right := c.compileValue(v.Right)
		rightLLVM := right.Value

		if left.PointerLevel > 0 {
			leftLLVM = c.contextBlock.NewLoad(leftLLVM)
		}

		if right.PointerLevel > 0 {
			rightLLVM = c.contextBlock.NewLoad(rightLLVM)
		}

		if !leftLLVM.Type().Equal(rightLLVM.Type()) {
			panic(fmt.Sprintf("Different types in operation: %T and %T", left, right))
		}

		switch leftLLVM.Type().GetName() {
		case "string":
			if v.Operator == parser.OP_ADD {
				leftLen := c.contextBlock.NewExtractValue(leftLLVM, []int64{0})
				rightLen := c.contextBlock.NewExtractValue(rightLLVM, []int64{0})
				sumLen := c.contextBlock.NewAdd(leftLen, rightLen)

				backingArray := c.contextBlock.NewAlloca(i8.LLVM())
				backingArray.NElems = sumLen

				// Copy left to new backing array
				c.contextBlock.NewCall(c.externalFuncs["strcpy"], backingArray, c.contextBlock.NewExtractValue(leftLLVM, []int64{1}))

				// Append right to backing array
				c.contextBlock.NewCall(c.externalFuncs["strcat"], backingArray, c.contextBlock.NewExtractValue(rightLLVM, []int64{1}))

				alloc := c.contextBlock.NewAlloca(typeConvertMap["string"].LLVM())

				// Save length of the string
				lenItem := c.contextBlock.NewGetElementPtr(alloc, constant.NewInt(0, i32.LLVM()), constant.NewInt(0, i32.LLVM()))
				c.contextBlock.NewStore(sumLen, lenItem)

				// Save i8* version of string
				strItem := c.contextBlock.NewGetElementPtr(alloc, constant.NewInt(0, i32.LLVM()), constant.NewInt(1, i32.LLVM()))
				c.contextBlock.NewStore(backingArray, strItem)

				return value.Value{
					Value:        c.contextBlock.NewLoad(alloc),
					Type:         types.String,
					PointerLevel: 0,
				}
			}

			panic("string does not implement operation " + v.Operator)
		}

		var opRes llvmValue.Value

		switch v.Operator {
		case parser.OP_ADD:
			opRes = c.contextBlock.NewAdd(leftLLVM, rightLLVM)
		case parser.OP_SUB:
			opRes = c.contextBlock.NewSub(leftLLVM, rightLLVM)
		case parser.OP_MUL:
			opRes = c.contextBlock.NewMul(leftLLVM, rightLLVM)
		case parser.OP_DIV:
			opRes = c.contextBlock.NewSDiv(leftLLVM, rightLLVM) // SDiv == Signed Division
		}

		return value.Value{
			Value:        opRes,
			Type:         left.Type,
			PointerLevel: 0,
		}

	case parser.NameNode:
		return c.varByName(v.Name)

	case parser.CallNode:
		var args []llvmValue.Value

		name, isNameNode := v.Function.(parser.NameNode)

		// len() functions
		if isNameNode && name.Name == "len" {
			arg := c.compileValue(v.Arguments[0])

			if arg.Type.Name() == "string" {
				f := c.funcByName("len_string")

				val := arg.Value
				if arg.PointerLevel > 0 {
					val = c.contextBlock.NewLoad(val)
				}

				return value.Value{
					Value:        c.contextBlock.NewCall(f.LlvmFunction, val),
					Type:         f.ReturnType,
					PointerLevel: 0,
				}
			}

			if arg.Type.Name() == "array" {
				if ptrType, ok := arg.Value.Type().(*llvmTypes.PointerType); ok {
					if arrayType, ok := ptrType.Elem.(*llvmTypes.ArrayType); ok {
						return value.Value{
							Value:        constant.NewInt(arrayType.Len, i64.LLVM()),
							Type:         i64,
							PointerLevel: 0,
						}
					}
				}
			}

			// if ptrType, ok := arg.Type().(*types.PointerType); ok {
			// 	if arrayType, ok := ptrType.Elem.(*types.ArrayType); ok {
			// 		return constant.NewInt(arrayType.Len, i64)
			// 	}
			// }
		}

		isExternal := false
		if isNameNode {
			_, isExternal = c.externalFuncs[name.Name]
		}

		for _, vv := range v.Arguments {
			treVal := c.compileValue(vv)
			val := treVal.Value

			// Convert %string* to i8* when calling external functions
			if isExternal {
				if treVal.Type.Name() == "string" {
					if treVal.PointerLevel > 0 {
						val = c.contextBlock.NewLoad(val)
					}
					args = append(args, c.contextBlock.NewExtractValue(val, []int64{1}))
					continue
				}

				if treVal.Type.Name() == "array" {
					if treVal.PointerLevel > 0 {
						val = c.contextBlock.NewLoad(val)
					}
					args = append(args, c.contextBlock.NewExtractValue(val, []int64{1}))
					continue
				}
			}

			if treVal.PointerLevel > 0 {
				val = c.contextBlock.NewLoad(val)
			}
			args = append(args, val)
		}

		var fn *types.Function

		if isNameNode {
			fn = c.funcByName(name.Name)
		} else {
			funcByVal := c.compileValue(v.Function)
			if checkIfFunc, ok := funcByVal.Type.(*types.Function); ok {
				fn = checkIfFunc
			} else if checkIfMethod, ok := funcByVal.Type.(*types.Method); ok {
				fn = checkIfMethod.Function
				var methodCallArgs []llvmValue.Value

				instance := funcByVal.Value
				if !checkIfMethod.PointerReceiver {
					instance = c.contextBlock.NewLoad(instance)
				}

				// Add instance as the first argument
				methodCallArgs = append(methodCallArgs, instance)
				methodCallArgs = append(methodCallArgs, args...)
				args = methodCallArgs
			} else {
				panic("expected function or method, got something else")
			}
		}

		// Call function and return the result
		return value.Value{
			Value:        c.contextBlock.NewCall(fn.LlvmFunction, args...),
			Type:         fn.ReturnType,
			PointerLevel: 0,
		}

	case parser.TypeCastNode:
		val := c.compileValue(v.Val)

		var current *llvmTypes.IntType
		var ok bool

		current, ok = val.Type.LLVM().(*llvmTypes.IntType)
		if !ok {
			panic("TypeCast origin must be int type")
		}

		targetType := parserTypeToType(v.Type)
		target, ok := targetType.LLVM().(*llvmTypes.IntType)
		if !ok {
			panic("TypeCast target must be int type")
		}

		llvmVal := val.Value
		if val.PointerLevel > 0 {
			llvmVal = c.contextBlock.NewLoad(llvmVal)
		}

		// Same size, nothing to do here
		if current.Size == target.Size {
			return val
		}

		res := c.contextBlock.NewAlloca(target)

		var changedSize llvmValue.Value

		if current.Size < target.Size {
			changedSize = c.contextBlock.NewSExt(llvmVal, target)
		} else {
			changedSize = c.contextBlock.NewTrunc(llvmVal, target)
		}

		c.contextBlock.NewStore(changedSize, res)

		return value.Value{
			Value:        res,
			Type:         targetType,
			PointerLevel: 1,
		}

	case parser.StructLoadElementNode:
		src := c.compileValue(v.Struct)

		if src.PointerLevel == 0 {
			// GetElementPtr only works on pointer types, and we don't have a pointer to our object.
			// Allocate it and use the pointer instead
			dst := c.contextBlock.NewAlloca(src.Type.LLVM())
			c.contextBlock.NewStore(src.Value, dst)
			src = value.Value{
				Value:        dst,
				Type:         src.Type,
				PointerLevel: src.PointerLevel + 1,
			}
		}

		// Check if it is a struct member
		if structType, ok := src.Type.(*types.Struct); ok {
			if memberIndex, ok := structType.MemberIndexes[v.ElementName]; ok {
				retVal := c.contextBlock.NewGetElementPtr(src.Value, constant.NewInt(0, i32.LLVM()), constant.NewInt(int64(memberIndex), i32.LLVM()))
				return value.Value{
					Type:         structType.Members[v.ElementName],
					Value:        retVal,
					PointerLevel: 1, // TODO
				}
			}
		}

		// Check if it's a method
		if method, ok := src.Type.GetMethod(v.ElementName); ok {
			return value.Value{
				Type:         method,
				Value:        src.Value,
				PointerLevel: 0,
			}
		}

		panic(fmt.Sprintf("%T internal error: no such type map indexing: %s", src, v.ElementName))

	case parser.SliceArrayNode:
		src := c.compileValue(v.Val)
		srcVal := src.Value

		var originalLength *ir.InstExtractValue

		// Get backing array from string type
		if src.PointerLevel > 0 {
			srcVal = c.contextBlock.NewLoad(srcVal)
		}
		if src.Type.Name() == "string" {
			originalLength = c.contextBlock.NewExtractValue(srcVal, []int64{0})
			srcVal = c.contextBlock.NewExtractValue(srcVal, []int64{1})
		}

		start := c.compileValue(v.Start)

		outsideOfLengthBr := c.contextBlock.Parent.NewBlock(getBlockName())
		c.panic(outsideOfLengthBr, "Substring start larger than len")
		outsideOfLengthBr.NewUnreachable()

		safeBlock := c.contextBlock.Parent.NewBlock(getBlockName())

		// Make sure that the offset is within the string length
		cmp := c.contextBlock.NewICmp(ir.IntUGE, start.Value, originalLength)
		c.contextBlock.NewCondBr(cmp, outsideOfLengthBr, safeBlock)

		c.contextBlock = safeBlock

		offset := safeBlock.NewGetElementPtr(srcVal, start.Value)

		var length llvmValue.Value
		if v.HasEnd {
			length = c.compileValue(v.End).Value
		} else {
			length = constant.NewInt(1, i64.LLVM())
		}

		dst := safeBlock.NewCall(c.externalFuncs["strndup"], offset, length)

		// Convert *i8 to %string
		alloc := safeBlock.NewAlloca(typeConvertMap["string"].LLVM())

		// Save length of the string
		lenItem := safeBlock.NewGetElementPtr(alloc, constant.NewInt(0, i32.LLVM()), constant.NewInt(0, i32.LLVM()))
		safeBlock.NewStore(constant.NewInt(100, i64.LLVM()), lenItem) // TODO

		// Save i8* version of string
		strItem := safeBlock.NewGetElementPtr(alloc, constant.NewInt(0, i32.LLVM()), constant.NewInt(1, i32.LLVM()))
		safeBlock.NewStore(dst, strItem)

		return value.Value{
			Value:        safeBlock.NewLoad(alloc),
			Type:         typeConvertMap["string"],
			PointerLevel: 0,
		}

	case parser.LoadArrayElement:
		arr := c.compileValue(v.Array)
		arrayValue := arr.Value

		index := c.compileValue(v.Pos)

		var runtimeLength llvmValue.Value
		var compileTimeLenght int64
		lengthKnownAtCompileTime := false
		lengthKnownAtRunTime := false
		arrIsString := false
		var retType types.Type

		// Length of array
		if ptrType, ok := arr.Value.Type().(*llvmTypes.PointerType); ok {
			if arrayType, ok := ptrType.Elem.(*llvmTypes.ArrayType); ok {
				compileTimeLenght = arrayType.Len
				lengthKnownAtCompileTime = true
				arrType := arr.Type.(*types.Array)
				retType = arrType.Type
			}
		}

		// Length of string
		if !lengthKnownAtCompileTime {
			// Get backing array from string type
			if arr.Type.Name() == "string" {
				if arr.PointerLevel > 0 {
					arrayValue = c.contextBlock.NewLoad(arrayValue)
				}

				runtimeLength = c.contextBlock.NewExtractValue(arrayValue, []int64{0})
				// Get backing array
				arrayValue = c.contextBlock.NewExtractValue(arrayValue, []int64{1})
				lengthKnownAtRunTime = true
				arrIsString = true
				retType = types.I8
			}
		}

		if !lengthKnownAtCompileTime && !lengthKnownAtRunTime {
			panic("unable to LoadArrayElement: could not calculate max length")
		}

		isCheckedAtCompileTime := false

		if lengthKnownAtCompileTime {
			if compileTimeLenght < 0 {
				compilePanic("index out of range")
			}

			if intType, ok := index.Value.(*constant.Int); ok {
				if intType.X.IsInt64() {
					isCheckedAtCompileTime = true

					if intType.X.Int64() > compileTimeLenght {
						compilePanic("index out of range")
					}
				}
			}
		}

		if !isCheckedAtCompileTime {
			outsideOfLengthBlock := c.contextBlock.Parent.NewBlock(getBlockName())
			c.panic(outsideOfLengthBlock, "index out of range")
			outsideOfLengthBlock.NewUnreachable()

			safeBlock := c.contextBlock.Parent.NewBlock(getBlockName())

			outOfRangeCmp := c.contextBlock.NewOr(
				c.contextBlock.NewICmp(ir.IntSLT, index.Value, constant.NewInt(0, i64.LLVM())),
				c.contextBlock.NewICmp(ir.IntSGE, index.Value, runtimeLength),
			)

			c.contextBlock.NewCondBr(outOfRangeCmp, outsideOfLengthBlock, safeBlock)

			c.contextBlock = safeBlock
		}

		var indicies []llvmValue.Value
		if !arrIsString {
			indicies = append(indicies, constant.NewInt(0, i64.LLVM()))
		}
		indicies = append(indicies, index.Value)

		return value.Value{
			Value:        c.contextBlock.NewGetElementPtr(arrayValue, indicies...),
			Type:         retType,
			PointerLevel: 1,
		}

	case parser.GetReferenceNode:
		return c.compileGetReferenceNode(v)
	case parser.DereferenceNode:
		return c.compileDereferenceNode(v)
	}

	panic("compileValue fail: " + fmt.Sprintf("%T: %+v", node, node))
}

func (c *compiler) panic(block *ir.BasicBlock, message string) {
	globMsg := c.module.NewGlobalDef(strings.NextStringName(), strings.Constant("runtime panic: "+message+"\n"))
	globMsg.IsConst = true
	block.NewCall(c.externalFuncs["printf"], strings.Toi8Ptr(block, globMsg))
	block.NewCall(c.externalFuncs["exit"], constant.NewInt(1, i32.LLVM()))
}

func compilePanic(message string) {
	fmt.Fprintf(os.Stderr, "compile panic: %s\n", message)
	os.Exit(1)
}
