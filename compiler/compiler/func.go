package compiler

import (
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	llvmTypes "github.com/llir/llvm/ir/types"
	llvmValue "github.com/llir/llvm/ir/value"

	"github.com/zegl/tre/compiler/compiler/internal"
	"github.com/zegl/tre/compiler/compiler/internal/pointer"
	"github.com/zegl/tre/compiler/compiler/name"
	"github.com/zegl/tre/compiler/compiler/types"
	"github.com/zegl/tre/compiler/compiler/value"
	"github.com/zegl/tre/compiler/parser"
)

func (c *Compiler) funcType(params, returnTypes []parser.TypeNode) (retType types.Type, treReturnTypes []types.Type, argTypes []*ir.Param, treParams []types.Type, isVariadicFunc bool, argumentReturnValuesCount int) {
	llvmParams := make([]*ir.Param, len(params))
	treParams = make([]types.Type, len(params))

	for k, par := range params {
		paramType := c.parserTypeToType(par)

		// Variadic arguments are converted into a slice
		// The function takes a slice as the argument, the caller has to convert
		// the arguments to a slice before calling
		if par.Variadic() {
			paramType = &types.Slice{
				Type:     paramType,
				LlvmType: internal.Slice(paramType.LLVM()),
			}
		}

		param := ir.NewParam(name.Var("p"), paramType.LLVM())

		if par.Variadic() {
			if k < len(params)-1 {
				panic("Only the last parameter can be varadic")
			}
			isVariadicFunc = true
		}

		llvmParams[k] = param
		treParams[k] = paramType
	}

	var funcRetType types.Type = types.Void

	// Amount of values returned via argument pointers
	// var argumentReturnValuesCount int
	// var treReturnTypes []types.Type

	// Use LLVM function return value if there's only one return value
	if len(returnTypes) == 1 {
		funcRetType = c.parserTypeToType(returnTypes[0])
		treReturnTypes = []types.Type{funcRetType}
	} else if len(returnTypes) > 0 {
		// Return values via argument pointers
		// The return values goes first
		var llvmReturnTypesParams []*ir.Param

		for _, ret := range returnTypes {
			t := c.parserTypeToType(ret)
			treReturnTypes = append(treReturnTypes, t)
			llvmReturnTypesParams = append(llvmReturnTypesParams, ir.NewParam(name.Var("ret"), llvmTypes.NewPointer(t.LLVM())))
		}

		// Add return values to the start
		treParams = append(treReturnTypes, treParams...)
		llvmParams = append(llvmReturnTypesParams, llvmParams...)

		argumentReturnValuesCount = len(returnTypes)
	}

	return funcRetType, treReturnTypes, llvmParams, treParams, isVariadicFunc, argumentReturnValuesCount
}

func (c *Compiler) compileDefineFuncNode(v *parser.DefineFuncNode) value.Value {
	var compiledName string

	if v.IsMethod {
		var methodOnType parser.TypeNode = v.MethodOnType

		if v.IsPointerReceiver {
			methodOnType = &parser.PointerTypeNode{ValueType: methodOnType}
		}

		// Add the type that we're a method on as the first argument
		v.Arguments = append([]*parser.NameNode{
			{
				Name: v.InstanceName,
				Type: methodOnType,
			},
		}, v.Arguments...)

		// Change the name of our function
		compiledName = c.currentPackageName + "_method_" + v.MethodOnType.TypeName + "_" + v.Name
	} else if v.IsNamed {
		compiledName = c.currentPackageName + "_" + v.Name
	} else {
		compiledName = c.currentPackageName + "_" + name.AnonFunc()
	}

	// Anonymous functions are closures, and take the closure environment as
	// the first parameter
	isClosure := !v.IsMethod && !v.IsNamed

	argTypes := make([]parser.TypeNode, len(v.Arguments))
	for k, v := range v.Arguments {
		argTypes[k] = v.Type
	}

	retTypes := make([]parser.TypeNode, len(v.ReturnValues))
	for k, v := range v.ReturnValues {
		retTypes[k] = v.Type
	}

	funcRetType, treReturnTypes, llvmParams, treParams, isVariadicFunc, argumentReturnValuesCount := c.funcType(argTypes, retTypes)

	var fn *ir.Func
	var entry *ir.Block
	var envParam *ir.Param

	if c.currentPackageName == "main" && v.Name == "main" {
		if len(v.ReturnValues) != 0 {
			panic("main func can not have a return type")
		}

		funcRetType = types.I32
		fn = c.mainFunc
		entry = fn.Blocks[0] // use already defined block
	} else if v.Name == "init" {
		fn = c.module.NewFunc(name.Var("init"), funcRetType.LLVM(), llvmParams...)
		entry = fn.NewBlock(name.Block())
		c.initGlobalsFunc.Blocks[0].NewCall(fn) // Setup call to init from the global init func
	} else if isClosure {
		envParam = ir.NewParam(name.Var("env"), types.ClosureEnvType)
		fn = c.module.NewFunc(compiledName, funcRetType.LLVM(), append([]*ir.Param{envParam}, llvmParams...)...)
		entry = fn.NewBlock(name.Block())
	} else {
		fn = c.module.NewFunc(compiledName, funcRetType.LLVM(), llvmParams...)
		entry = fn.NewBlock(name.Block())
	}

	// The function type without the closure environment
	var funcType llvmTypes.Type = fn.Type()
	if isClosure {
		paramTypes := make([]llvmTypes.Type, len(llvmParams))
		for i, p := range llvmParams {
			paramTypes[i] = p.Type()
		}
		funcType = llvmTypes.NewPointer(llvmTypes.NewFunc(funcRetType.LLVM(), paramTypes...))
	}

	typesFunc := &types.Function{
		FuncType:       funcType,
		LlvmReturnType: funcRetType,
		ReturnTypes:    treReturnTypes,
		IsVariadic:     isVariadicFunc,
		ArgumentTypes:  treParams,
	}

	// Save as a method on the type
	if v.IsMethod {
		if t, ok := c.currentPackage.GetPkgType(v.MethodOnType.TypeName, true); ok {
			t.AddMethod(v.Name, &types.Method{
				Function:        typesFunc,
				LlvmFunction:    fn,
				PointerReceiver: v.IsPointerReceiver,
				MethodName:      v.Name,
			})
		} else {
			panic("save method on type failed")
		}

		// Make this method available in interfaces via a jump function
		typesFunc.JumpFunction = c.compileInterfaceMethodJump(fn)
	} else if v.IsNamed {
		c.currentPackage.DefinePkgVar(v.Name, value.Value{
			Type:  typesFunc,
			Value: fn,
		})
	}

	// Create the environment of captured variables in the enclosing function
	var env llvmValue.Value = constant.NewNull(types.ClosureEnvType)
	var envType *llvmTypes.StructType
	var capturedVals []value.Value

	if len(v.Captures) > 0 {
		fieldTypes := make([]llvmTypes.Type, len(v.Captures))
		for i, captured := range v.Captures {
			val := c.lookupName(&parser.NameNode{Name: captured})
			capturedVals = append(capturedVals, val)
			fieldTypes[i] = val.Value.Type()
		}

		// Variables are captured by reference, the environment contains
		// pointers to the variables. Values that are not variables (such
		// as constants) are captured by value.
		envType = llvmTypes.NewStruct(fieldTypes...)
		envPtr := c.allocVar(envType, true)
		envPtr.SetName(name.Var("closure-env"))

		for i, val := range capturedVals {
			fieldPtr := c.contextBlock.NewGetElementPtr(envType, envPtr, constant.NewInt(llvmTypes.I32, 0), constant.NewInt(llvmTypes.I32, int64(i)))
			c.contextBlock.NewStore(val.Value, fieldPtr)
		}

		env = c.contextBlock.NewBitCast(envPtr, types.ClosureEnvType)
	}

	prevContextFunc := c.contextFunc
	prevContextBlock := c.contextBlock
	prevContextFuncScope := c.contextFuncScope

	// The function does not share any context with the enclosing function
	prevContextLoopBreak := c.contextLoopBreak
	prevContextLoopContinue := c.contextLoopContinue
	prevContextCondAfter := c.contextCondAfter
	prevContextAssignDest := c.contextAssignDest
	prevContextAlloc := c.contextAlloc
	c.contextLoopBreak = nil
	c.contextLoopContinue = nil
	c.contextCondAfter = nil
	c.contextAssignDest = nil
	c.contextAlloc = nil

	// Restored with defer, so that the context of the enclosing function is
	// intact when unwinding from compilation errors
	defer func() {
		c.contextFunc = prevContextFunc
		c.contextBlock = prevContextBlock
		c.contextFuncScope = prevContextFuncScope

		c.contextLoopBreak = prevContextLoopBreak
		c.contextLoopContinue = prevContextLoopContinue
		c.contextCondAfter = prevContextCondAfter
		c.contextAssignDest = prevContextAssignDest
		c.contextAlloc = prevContextAlloc
	}()

	c.contextFunc = typesFunc
	c.contextBlock = entry
	c.pushVariablesStack()
	c.contextFuncScope = len(c.contextBlockVariables) - 1

	// Load the captured variables from the environment.
	// Arguments are added after this, and can shadow captured variables.
	if len(capturedVals) > 0 {
		envPtr := entry.NewBitCast(envParam, llvmTypes.NewPointer(envType))

		for i, val := range capturedVals {
			fieldPtr := entry.NewGetElementPtr(envType, envPtr, constant.NewInt(llvmTypes.I32, 0), constant.NewInt(llvmTypes.I32, int64(i)))
			loaded := entry.NewLoad(envType.Fields[i], fieldPtr)
			loaded.SetName(name.Var(v.Captures[i]))

			c.setVar(v.Captures[i], value.Value{
				Type:       val.Type,
				Value:      loaded,
				IsVariable: val.IsVariable,
			})
		}
	}

	// Push to the return values stack
	if argumentReturnValuesCount > 0 {
		var retVals []value.Value

		for i, retType := range treParams[:argumentReturnValuesCount] {
			retVals = append(retVals, value.Value{
				Value:      llvmParams[i],
				Type:       retType,
				IsVariable: true,
			})
		}

		c.contextFuncRetVals = append(c.contextFuncRetVals, retVals)
	}

	// Save all parameters in the block mapping
	for i, param := range llvmParams {
		var paramName string
		var dataType types.Type
		var isVariable bool

		// Named return values
		if i < argumentReturnValuesCount {
			paramName = v.ReturnValues[i].Name
			dataType = treReturnTypes[i]
			isVariable = true
		} else {
			paramName = v.Arguments[i-argumentReturnValuesCount].Name
			dataType = treParams[i-argumentReturnValuesCount]
		}

		// Arguments that are captured by closures are moved to the heap
		escapes := i >= argumentReturnValuesCount && v.EscapingArguments[paramName]

		// Structs needs to be pointer-allocated
		if _, isStruct := param.Type().(*llvmTypes.StructType); isStruct || escapes {
			paramPtr := c.allocVar(dataType.LLVM(), escapes)
			paramPtr.SetName(name.Var("paramPtr"))
			entry.NewStore(param, paramPtr)

			c.setVar(paramName, value.Value{
				Value:      paramPtr,
				Type:       dataType,
				IsVariable: true,
			})

			continue
		}

		c.setVar(paramName, value.Value{
			Value:      param,
			Type:       dataType,
			IsVariable: isVariable,
		})
	}

	// Single return value (not via parameters)
	// Add to variable block
	if len(v.ReturnValues) == 1 {
		r := v.ReturnValues[0]
		all := c.contextBlock.NewAlloca(funcRetType.LLVM())
		retVar := value.Value{
			Value:      all,
			Type:       funcRetType,
			IsVariable: true,
		}
		c.setVar(r.Name, retVar)
		c.contextFuncRetVals = append(c.contextFuncRetVals, []value.Value{retVar})
	}

	c.compile(v.Body)

	// Return void if there is no return type explicitly set
	if len(v.ReturnValues) == 0 {
		c.contextBlock.NewRet(nil)
	} else {
		// Pop return variables context
		c.contextFuncRetVals = c.contextFuncRetVals[0 : len(c.contextFuncRetVals)-1]
	}

	// Return 0 by default in main func
	if v.Name == "main" {
		c.contextBlock.NewRet(constant.NewInt(llvmTypes.I32, 0))
	}

	c.popVariablesStack()

	if isClosure {
		// Created in the enclosing function
		c.contextBlock = prevContextBlock
		return c.closureValue(typesFunc, fn, env)
	}

	return value.Value{
		Type:  typesFunc,
		Value: fn,
	}
}

// closureValue creates a closure value from a function and its environment
func (c *Compiler) closureValue(fnType *types.Function, fn, env llvmValue.Value) value.Value {
	closureType := fnType.LLVM().(*llvmTypes.StructType)

	// Use a constant when possible, as closures can be created outside of functions
	fnConst, fnIsConst := fn.(constant.Constant)
	envConst, envIsConst := env.(constant.Constant)
	if fnIsConst && envIsConst {
		return value.Value{
			Type:  fnType,
			Value: constant.NewStruct(closureType, fnConst, envConst),
		}
	}

	closure := c.contextBlock.NewInsertValue(constant.NewZeroInitializer(closureType), fn, 0)
	return value.Value{
		Type:  fnType,
		Value: c.contextBlock.NewInsertValue(closure, env, 1),
	}
}

// funcToClosure converts a named function to a closure, so that it can be
// used as a func value. The closure calls a wrapper function that discards the
// closure environment.
func (c *Compiler) funcToClosure(v value.Value) value.Value {
	fnType := v.Type.(*types.Function)
	fn := v.Value.(*ir.Func)

	if fnType.IsExternal {
		compilePanic("external functions can not be used as values")
	}

	wrapper, ok := c.closureWrappers[fn]
	if !ok {
		params := []*ir.Param{ir.NewParam("env", types.ClosureEnvType)}
		var args []llvmValue.Value
		for _, p := range fn.Params {
			param := ir.NewParam("", p.Type())
			params = append(params, param)
			args = append(args, param)
		}

		wrapper = c.module.NewFunc(fn.Name()+"_closure", fn.Sig.RetType, params...)
		block := wrapper.NewBlock(name.Block())
		res := block.NewCall(fn, args...)

		if _, ok := fn.Sig.RetType.(*llvmTypes.VoidType); ok {
			block.NewRet(nil)
		} else {
			block.NewRet(res)
		}

		c.closureWrappers[fn] = wrapper
	}

	return c.closureValue(fnType, wrapper, constant.NewNull(types.ClosureEnvType))
}

func (c *Compiler) compileInterfaceMethodJump(targetFunc *ir.Func) *ir.Func {
	// Copy parameter types so that we can modify them
	params := make([]*ir.Param, len(targetFunc.Sig.Params))
	for i, p := range targetFunc.Params {
		params[i] = ir.NewParam("", p.Type())
	}

	originalType := targetFunc.Params[0].Type()
	_, isPointerType := originalType.(*llvmTypes.PointerType)
	if !isPointerType {
		originalType = llvmTypes.NewPointer(originalType)
	}

	// Replace the first parameter type with an *i8
	// Will be bitcasted later to the target type
	params[0] = ir.NewParam("unsafe-ptr", llvmTypes.NewPointer(llvmTypes.I8))

	fn := c.module.NewFunc(targetFunc.Name()+"_jump", targetFunc.Sig.RetType, params...)
	block := fn.NewBlock(name.Block())

	var bitcasted llvmValue.Value = block.NewBitCast(params[0], originalType)

	// TODO: Don't do this if the method has a pointer receiver
	if !isPointerType {
		bitcasted = block.NewLoad(pointer.ElemType(bitcasted), bitcasted)
	}

	callArgs := []llvmValue.Value{bitcasted}
	for _, p := range params[1:] {
		callArgs = append(callArgs, p)
	}

	resVal := block.NewCall(targetFunc, callArgs...)

	if _, ok := targetFunc.Sig.RetType.(*llvmTypes.VoidType); ok {
		block.NewRet(nil)
	} else {
		block.NewRet(resVal)
	}

	return fn
}

func (c *Compiler) compileReturnNode(v *parser.ReturnNode) {
	// Single variable return
	if len(v.Vals) == 1 {
		// Set value and jump to return block
		val := c.compileValue(v.Vals[0])

		// Type cast if necessary
		val = c.valueToInterfaceValue(val, c.contextFunc.LlvmReturnType)

		if val.IsVariable {
			c.contextBlock.NewRet(c.contextBlock.NewLoad(pointer.ElemType(val.Value), val.Value))
			return
		}

		c.contextBlock.NewRet(val.Value)
		return
	}

	// Multiple value returns
	if len(v.Vals) > 1 {
		for i, val := range v.Vals {
			compVal := c.compileValue(val)

			// TODO: Type cast if necessary
			// compVal = c.valueToInterfaceValue(compVal, c.contextFunc.ReturnType)

			retVal := internal.LoadIfVariable(c.contextBlock, compVal)

			// Assign to ptr
			retValPtr := c.contextFuncRetVals[len(c.contextFuncRetVals)-1][i]

			c.contextBlock.NewStore(retVal, retValPtr.Value)
		}

		c.contextBlock.NewRet(nil)
		return
	}

	// Naked return, func has one named return variable
	if len(v.Vals) == 0 {
		retVals := c.contextFuncRetVals[len(c.contextFuncRetVals)-1]
		if len(retVals) == 1 {
			val := internal.LoadIfVariable(c.contextBlock, retVals[0])
			c.contextBlock.NewRet(val)
			return
		}
	}

	// Return void in LLVM function
	c.contextBlock.NewRet(nil)
}

func (c *Compiler) compileCallNode(v *parser.CallNode) value.Value {
	var args []value.Value

	name, isNameNode := v.Function.(*parser.NameNode)

	if isNameNode {
		switch name.Name {
		case "len":
			return c.lenFuncCall(v)
		case "cap":
			return c.capFuncCall(v)
		case "append":
			return c.appendFuncCall(v)
		case "print":
			return c.printFuncCall(v)
		}
	}

	var fnType *types.Function
	var fn llvmValue.Value

	// The environment of the closure that is called, is nil when calling a
	// named function directly
	var closureEnv llvmValue.Value

	// Look up named functions without converting them to closures
	var funcByVal value.Value
	if isNameNode {
		funcByVal = c.lookupName(name)
	} else {
		funcByVal = c.compileValue(v.Function)
	}

	if checkIfFunc, ok := funcByVal.Type.(*types.Function); ok {
		fnType = checkIfFunc
		if directFn, ok := funcByVal.Value.(*ir.Func); ok && !funcByVal.IsVariable {
			fn = directFn
		} else {
			closure := internal.LoadIfVariable(c.contextBlock, funcByVal)
			fn = c.contextBlock.NewExtractValue(closure, 0)
			closureEnv = c.contextBlock.NewExtractValue(closure, 1)
		}
	} else if checkIfMethod, ok := funcByVal.Type.(*types.Method); ok {
		fnType = checkIfMethod.Function
		fn = checkIfMethod.LlvmFunction

		var methodCallArgs []value.Value

		// Should be loaded if the method is not a pointer receiver
		funcByVal.IsVariable = !checkIfMethod.PointerReceiver

		// Add instance as the first argument
		methodCallArgs = append(methodCallArgs, funcByVal)
		methodCallArgs = append(methodCallArgs, args...)
		args = methodCallArgs
	} else if ifaceMethod, ok := funcByVal.Type.(*types.InterfaceMethod); ok {

		ifaceInstance := c.contextBlock.NewGetElementPtr(pointer.ElemType(funcByVal.Value), funcByVal.Value, constant.NewInt(llvmTypes.I32, 0), constant.NewInt(llvmTypes.I32, 0))
		ifaceInstanceLoad := c.contextBlock.NewLoad(pointer.ElemType(ifaceInstance), ifaceInstance)

		// Add instance as the first argument
		var methodCallArgs []value.Value
		methodCallArgs = append(methodCallArgs, value.Value{
			Value: ifaceInstanceLoad,
		})
		methodCallArgs = append(methodCallArgs, args...)
		args = methodCallArgs

		var returnType types.Type
		if len(ifaceMethod.ReturnTypes) > 0 {
			returnType = ifaceMethod.ReturnTypes[0]
		} else {
			returnType = types.Void
		}

		fnType = &types.Function{
			// TODO: We probably need to add more fields here?
			FuncType:       ifaceMethod.LlvmJumpFunction.Type(),
			LlvmReturnType: returnType,
		}
		fn = ifaceMethod.LlvmJumpFunction
	} else {
		panic("expected function or method, got something else")
	}

	// If the last argument is a slice that is "de variadicified"
	// Eg: foo...
	// When this is the case we don't have to convert the arguments to a slice when calling the func
	lastIsVariadicSlice := false

	// Compile all values
	for _, vv := range v.Arguments {
		if devVar, ok := vv.(*parser.DeVariadicSliceNode); ok {
			lastIsVariadicSlice = true
			args = append(args, c.compileValue(devVar.Item))
			continue
		}
		args = append(args, c.compileValue(vv))
	}

	// Convert variadic arguments to a slice when needed
	if fnType.IsVariadic && !lastIsVariadicSlice {
		// Only the last argument can be variadic
		variadicArgIndex := len(fnType.ArgumentTypes) - 1
		variadicType := fnType.ArgumentTypes[variadicArgIndex].(*types.Slice)

		// Convert last argument to a slice.
		variadicSlice := c.compileInitializeSliceWithValues(variadicType.Type, args[variadicArgIndex:]...)

		// Remove "pre-sliceified" arguments from the list of arguments
		args = args[0:variadicArgIndex]
		args = append(args, variadicSlice)
	}

	// Convert all values to LLVM values
	// Load the variable if needed
	llvmArgs := make([]llvmValue.Value, len(args))
	for i, v := range args {

		// Convert type to interface type if needed
		if len(fnType.ArgumentTypes) > i {
			v = c.valueToInterfaceValue(v, fnType.ArgumentTypes[i])
		}

		val := internal.LoadIfVariable(c.contextBlock, v)

		// Convert strings and arrays to i8* when calling external functions
		if fnType.IsExternal {
			if v.Type.Name() == "string" {
				llvmArgs[i] = c.contextBlock.NewExtractValue(val, 1)
				continue
			}

			if v.Type.Name() == "array" {
				llvmArgs[i] = c.contextBlock.NewExtractValue(val, 1)
				continue
			}
		}

		llvmArgs[i] = val
	}

	// Functions with multiple return values are using pointers via arguments
	// Alloc the values here and add pointers to the list of arguments
	var multiValues []value.Value
	if len(fnType.ReturnTypes) > 1 {
		var retValAllocas []llvmValue.Value

		for _, retType := range fnType.ReturnTypes {
			alloca := c.contextBlock.NewAlloca(retType.LLVM())
			retValAllocas = append(retValAllocas, alloca)

			multiValues = append(multiValues, value.Value{
				Type:       retType,
				Value:      alloca,
				IsVariable: true,
			})
		}

		// Add to start of argument list
		llvmArgs = append(retValAllocas, llvmArgs...)
	}

	// The closure environment is always the first argument
	if closureEnv != nil {
		llvmArgs = append([]llvmValue.Value{closureEnv}, llvmArgs...)
	}

	funcCallRes := c.contextBlock.NewCall(fn, llvmArgs...)

	// 0 or 1 return variables
	if len(fnType.ReturnTypes) < 2 {
		return value.Value{
			Value: funcCallRes,
			Type:  fnType.LlvmReturnType,
		}
	}

	// 2 or more return variables
	return value.Value{
		Type:        &types.MultiValue{Types: fnType.ReturnTypes},
		MultiValues: multiValues,
	}
}
