package compiler

import (
	"fmt"
	"runtime"
	"runtime/debug"

	"github.com/zegl/tre/compiler/compiler/internal"
	"github.com/zegl/tre/compiler/compiler/strings"
	"github.com/zegl/tre/compiler/compiler/types"
	"github.com/zegl/tre/compiler/compiler/value"
	"github.com/zegl/tre/compiler/parser"

	"errors"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	llvmTypes "github.com/llir/llvm/ir/types"
	llvmValue "github.com/llir/llvm/ir/value"
)

type Compiler struct {
	module *ir.Module

	// functions provided by the OS, such as printf
	externalFuncs map[string]*ir.Function

	// functions provided by the language, such as println
	globalFuncs map[string]*types.Function

	packages           map[string]*types.PackageInstance
	currentPackage     *types.PackageInstance
	currentPackageName string

	contextFunc           *ir.Function
	contextBlock          *ir.BasicBlock
	contextBlockVariables map[string]value.Value

	// What a break or continue should resolve to
	contextLoopBreak    []*ir.BasicBlock
	contextLoopContinue []*ir.BasicBlock

	stringConstants map[string]*ir.Global
}

var (
	i8  = types.I8
	i32 = types.I32
	i64 = types.I64
)

func NewCompiler() *Compiler {
	c := &Compiler{
		module:        ir.NewModule(),
		externalFuncs: make(map[string]*ir.Function),
		globalFuncs:   make(map[string]*types.Function),

		packages: make(map[string]*types.PackageInstance),

		contextLoopBreak:    make([]*ir.BasicBlock, 0),
		contextLoopContinue: make([]*ir.BasicBlock, 0),

		stringConstants: make(map[string]*ir.Global),
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
	case "windows":
		targetTriple[1] = "pc-windows"
	default:
		panic("unsupported GOOS: " + runtime.GOOS)
	}

	c.module.TargetTriple = fmt.Sprintf("%s-%s", targetTriple[0], targetTriple[1])

	return c
}

func (c *Compiler) Compile(root parser.PackageNode) (err error) {
	defer func() {
		if r := recover(); r != nil {
			// Compile time panics, that are not errors in the compiler
			if _, ok := r.(Panic); ok {
				err = errors.New(fmt.Sprint(r))
				return
			}

			// Bugs in the compiler
			err = fmt.Errorf("%s\n\nInternal compiler stacktrace:\n%s",
				fmt.Sprint(r),
				string(debug.Stack()),
			)
		}
	}()

	c.currentPackage = &types.PackageInstance{
		Funcs: make(map[string]*types.Function),
	}
	c.currentPackageName = root.Name
	c.packages[c.currentPackageName] = c.currentPackage

	for _, fileNode := range root.Files {
		c.compile(fileNode.Instructions)
	}

	return
}

func (c *Compiler) GetIR() string {
	return fmt.Sprintln(c.module)
}

func (c *Compiler) addExternal() {
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

func (c *Compiler) addGlobal() {
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

func (c *Compiler) compile(instructions []parser.Node) {
	for _, i := range instructions {
		switch v := i.(type) {
		case parser.ConditionNode:
			cond := c.compileCondition(v.Cond)

			afterBlock := c.contextFunc.NewBlock(getBlockName() + "-after")
			trueBlock := c.contextFunc.NewBlock(getBlockName() + "-true")
			falseBlock := afterBlock

			if len(v.False) > 0 {
				falseBlock = c.contextFunc.NewBlock(getBlockName() + "-false")
			}

			c.contextBlock.NewCondBr(cond.Value, trueBlock, falseBlock)

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
				compiledName = c.currentPackageName + "_method_" + v.MethodOnType.TypeName + "_" + v.Name
			} else {
				compiledName = c.currentPackageName + "_" + v.Name
			}

			params := make([]*llvmTypes.Param, len(v.Arguments))
			for k, par := range v.Arguments {
				params[k] = ir.NewParam(par.Name, parserTypeToType(par.Type).LLVM())
			}

			var funcRetType types.Type = types.Void

			if len(v.ReturnValues) == 1 {
				funcRetType = parserTypeToType(v.ReturnValues[0].Type)
			}

			if c.currentPackageName == "main" && v.Name == "main" {
				if len(v.ReturnValues) != 0 {
					panic("main func can not have a return type")
				}

				funcRetType = types.I32
				compiledName = "main"
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
				c.currentPackage.Funcs[v.Name] = typesFunc
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
				c.contextBlock.NewRet(c.contextBlock.NewLoad(val.Value))
				break
			}

			c.contextBlock.NewRet(val.Value)
			break

		case parser.AllocNode:

			// Allocate from type
			if typeNode, ok := v.Val.(parser.TypeNode); ok {
				treType := parserTypeToType(typeNode)

				alloc := c.contextBlock.NewAlloca(treType.LLVM())
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
				llvmVal = c.contextBlock.NewLoad(llvmVal)
			}

			alloc := c.contextBlock.NewAlloca(llvmVal.Type())
			alloc.SetName(v.Name)
			c.contextBlock.NewStore(llvmVal, alloc)

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
					alloc := c.contextBlock.NewAlloca(parserTypeToType(singleTypeNode).LLVM())
					c.contextBlock.NewStore(c.contextBlock.NewLoad(alloc), llvmDst)
					break
				}

				panic("AssignNode from non TypeNode is not allowed")
			}

			// Allocate from value
			val := c.compileValue(v.Val)
			llvmV := val.Value

			if val.PointerLevel > 0 {
				llvmV = c.contextBlock.NewLoad(llvmV)
			}

			c.contextBlock.NewStore(llvmV, llvmDst)
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

		case parser.ImportNode:
			// NOOP
			break

		default:
			c.compileValue(v)
			break
		}
	}
}

func (c *Compiler) funcByName(name string) *types.Function {
	if f, ok := c.globalFuncs[name]; ok {
		return f
	}

	// Function in the current package
	if f, ok := c.currentPackage.Funcs[name]; ok {
		return f
	}

	panic("funcByName: no such func: " + name)
}

func (c *Compiler) varByName(name string) value.Value {
	// Named variable in this block?
	if val, ok := c.contextBlockVariables[name]; ok {
		return val
	}

	// Imported package?
	if pkg, ok := c.packages[name]; ok {
		return value.Value{
			Type: pkg,
		}
	}

	panic("undefined variable: " + name)
}

func (c *Compiler) compileValue(node parser.Node) value.Value {
	switch v := node.(type) {

	case parser.ConstantNode:
		return c.compileConstantNode(v)

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

			if arg.Type.Name() == "slice" {
				val := arg.Value
				val = c.contextBlock.NewLoad(val)

				// TODO: Why is a double load needed?
				if _, ok := val.Type().(*llvmTypes.PointerType); ok {
					val = c.contextBlock.NewLoad(val)
				}

				return value.Value{
					Value:        c.contextBlock.NewExtractValue(val, []int64{0}),
					Type:         i64,
					PointerLevel: 0,
				}
			}
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

		if packageRef, ok := src.Type.(*types.PackageInstance); ok {
			if f, ok := packageRef.Funcs[v.ElementName]; ok {
				return value.Value{
					Type: f,
				}
			}

			panic(fmt.Sprintf("Package %s has no such method %s", packageRef.Name(), v.ElementName))
		}

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

		if _, ok := src.Type.(*types.StringType); ok {
			return c.compileSubstring(src, v)
		}

		return c.compileSliceArray(src, v)

	case parser.LoadArrayElement:
		arr := c.compileValue(v.Array)
		arrayValue := arr.Value

		index := c.compileValue(v.Pos)
		indexVal := index.Value
		if index.PointerLevel > 0 {
			indexVal = c.contextBlock.NewLoad(indexVal)
		}

		var runtimeLength llvmValue.Value
		var compileTimeLenght int64
		lengthKnownAtCompileTime := false
		lengthKnownAtRunTime := false
		isLlvmArrayBased := false
		var retType types.Type

		// Length of array
		if arr, ok := arr.Type.(*types.Array); ok {
			if arrayType, ok := arr.LlvmType.(*llvmTypes.ArrayType); ok {
				compileTimeLenght = arrayType.Len
				lengthKnownAtCompileTime = true
				retType = arr.Type
				isLlvmArrayBased = true
			}
		}

		// Length of slice
		if slice, ok := arr.Type.(*types.Slice); ok {
			lengthKnownAtCompileTime = false
			lengthKnownAtRunTime = true

			retType = slice.Type

			// TODO: Figure out if this really is needed
			arrayValue = c.contextBlock.NewLoad(arrayValue)
			arrayValue = c.contextBlock.NewLoad(arrayValue)

			sliceValue := arrayValue

			// Length of the slice
			runtimeLength = c.contextBlock.NewExtractValue(sliceValue, []int64{0})

			// Add offset to indexVal
			backingArrayOffset := c.contextBlock.NewExtractValue(sliceValue, []int64{2})
			indexVal = c.contextBlock.NewAdd(indexVal, backingArrayOffset)

			// Add offset to runtimeLength
			runtimeLength = c.contextBlock.NewAdd(runtimeLength, backingArrayOffset)

			// Backing array
			arrayValue = c.contextBlock.NewExtractValue(sliceValue, []int64{3})
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
				retType = types.I8
				isLlvmArrayBased = false
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
			outsideOfLengthBlock := c.contextBlock.Parent.NewBlock(getBlockName() + "-array-index-out-of-range")
			c.panic(outsideOfLengthBlock, "index out of range")
			outsideOfLengthBlock.NewUnreachable()

			safeBlock := c.contextBlock.Parent.NewBlock(getBlockName() + "-after-array-index-check")

			var runtimeOrCompiletimeCmp *ir.InstICmp
			if lengthKnownAtCompileTime {
				runtimeOrCompiletimeCmp = c.contextBlock.NewICmp(ir.IntSGE, indexVal, constant.NewInt(compileTimeLenght, i32.LLVM()))
			} else {
				runtimeOrCompiletimeCmp = c.contextBlock.NewICmp(ir.IntSGE, indexVal, runtimeLength)
			}

			outOfRangeCmp := c.contextBlock.NewOr(
				c.contextBlock.NewICmp(ir.IntSLT, indexVal, constant.NewInt(0, i64.LLVM())),
				runtimeOrCompiletimeCmp,
			)

			c.contextBlock.NewCondBr(outOfRangeCmp, outsideOfLengthBlock, safeBlock)

			c.contextBlock = safeBlock
		}

		var indicies []llvmValue.Value
		if isLlvmArrayBased {
			indicies = append(indicies, constant.NewInt(0, i64.LLVM()))
		}
		indicies = append(indicies, indexVal)

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

func (c *Compiler) panic(block *ir.BasicBlock, message string) {
	globMsg := c.module.NewGlobalDef(strings.NextStringName(), strings.Constant("runtime panic: "+message+"\n"))
	globMsg.IsConst = true
	block.NewCall(c.externalFuncs["printf"], strings.Toi8Ptr(block, globMsg))
	block.NewCall(c.externalFuncs["exit"], constant.NewInt(1, i32.LLVM()))
}

type Panic string

func compilePanic(message string) {
	panic(Panic(fmt.Sprintf("compile panic: %s\n", message)))
}
