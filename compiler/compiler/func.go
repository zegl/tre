package compiler

import (
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	llvmTypes "github.com/llir/llvm/ir/types"
	llvmValue "github.com/llir/llvm/ir/value"
	"github.com/zegl/tre/compiler/compiler/internal"
	"github.com/zegl/tre/compiler/compiler/types"
	"github.com/zegl/tre/compiler/compiler/value"
	"github.com/zegl/tre/compiler/parser"
)

func (c *Compiler) compileDefineFuncNode(v parser.DefineFuncNode) {
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

	llvmParams := make([]*llvmTypes.Param, len(v.Arguments))
	treParams := make([]types.Type, len(v.Arguments))

	isVariadicFunc := false

	for k, par := range v.Arguments {
		paramType := parserTypeToType(par.Type)

		// Variadic arguments are converted into a slice
		// The function takes a slice as the argument, the caller has to convert
		// the arguments to a slice before calling
		if par.Type.Variadic() {
			paramType = &types.Slice{
				Type:     paramType,
				LlvmType: internal.Slice(paramType.LLVM()),
			}
		}

		param := ir.NewParam(par.Name, paramType.LLVM())

		// TODO: Should only be possible on the last argument
		if par.Type.Variadic() {
			isVariadicFunc = true
		}

		llvmParams[k] = param
		treParams[k] = paramType
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
	fn := c.module.NewFunction(compiledName, funcRetType.LLVM(), llvmParams...)

	typesFunc := &types.Function{
		LlvmFunction:  fn,
		ReturnType:    funcRetType,
		FunctionName:  v.Name,
		IsVariadic:    isVariadicFunc,
		ArgumentTypes: treParams,
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
	for i, param := range llvmParams {
		paramName := v.Arguments[i].Name
		dataType := treParams[i]

		// Structs needs to be pointer-allocated
		if _, ok := param.Type().(*llvmTypes.StructType); ok {
			paramPtr := entry.NewAlloca(dataType.LLVM())
			entry.NewStore(param, paramPtr)

			c.contextBlockVariables[paramName] = value.Value{
				Value:      paramPtr,
				Type:       dataType,
				IsVariable: true,
			}

			continue
		}

		c.contextBlockVariables[paramName] = value.Value{
			Value:      param,
			Type:       dataType,
			IsVariable: false,
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
}

func (c *Compiler) compileReturnNode(v parser.ReturnNode) {
	// Set value and jump to return block
	val := c.compileValue(v.Val)

	if val.IsVariable {
		c.contextBlock.NewRet(c.contextBlock.NewLoad(val.Value))
		return
	}

	c.contextBlock.NewRet(val.Value)
}

func (c *Compiler) compileCallNode(v parser.CallNode) value.Value {
	var args []value.Value

	name, isNameNode := v.Function.(parser.NameNode)

	// len() functions
	if isNameNode && name.Name == "len" {
		return c.lenFuncCall(v)
	}

	// cap() function
	if isNameNode && name.Name == "cap" {
		return c.capFuncCall(v)
	}

	// append() function
	if isNameNode && name.Name == "append" {
		return c.appendFuncCall(v)
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
			var methodCallArgs []value.Value

			// Should be loaded if the method is not a pointer receiver
			funcByVal.IsVariable = !checkIfMethod.PointerReceiver

			// Add instance as the first argument
			methodCallArgs = append(methodCallArgs, funcByVal)
			methodCallArgs = append(methodCallArgs, args...)
			args = methodCallArgs
		} else {
			panic("expected function or method, got something else")
		}
	}

	// If the last argument is a slice that is "de variadicified"
	// Eg: foo...
	// When this is the case we don't have to convert the arguments to a slice when calling the func
	lastIsVariadicSlice := false

	// Compile all values
	for _, vv := range v.Arguments {
		if devVar, ok := vv.(parser.DeVariadicSliceNode); ok {
			lastIsVariadicSlice = true
			args = append(args, c.compileValue(devVar.Item))
			continue
		}
		args = append(args, c.compileValue(vv))
	}

	// Convert variadic arguments to a slice when needed
	if fn.IsVariadic && !lastIsVariadicSlice {
		// Only the last argument can be variadic
		variadicArgIndex := len(fn.ArgumentTypes) - 1
		variadicType := fn.ArgumentTypes[variadicArgIndex].(*types.Slice)

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
		val := v.Value

		if v.IsVariable {
			val = c.contextBlock.NewLoad(val)
		}

		// Convert strings and arrays to i8* when calling external functions
		if fn.IsExternal {
			if v.Type.Name() == "string" {
				llvmArgs[i] = c.contextBlock.NewExtractValue(val, []int64{1})
				continue
			}

			if v.Type.Name() == "array" {
				llvmArgs[i] = c.contextBlock.NewExtractValue(val, []int64{1})
				continue
			}
		}

		if len(fn.ArgumentTypes) > i {
			fnArgumentType := fn.ArgumentTypes[i]
			if _, isInterface := fnArgumentType.(*types.Interface); isInterface {
				// Convert to pointer variable
				if !v.IsVariable {
					ptrAlloca := c.contextBlock.NewAlloca(v.Type.LLVM())
					c.contextBlock.NewStore(val, ptrAlloca)
					val = ptrAlloca
				}

				ifaceStruct := c.contextBlock.NewAlloca(internal.Interface())
				dataPtr := c.contextBlock.NewGetElementPtr(ifaceStruct, constant.NewInt(0, i32.LLVM()), constant.NewInt(0, i32.LLVM()))
				bitcastedVal := c.contextBlock.NewBitCast(val, llvmTypes.NewPointer(llvmTypes.I8))
				c.contextBlock.NewStore(bitcastedVal, dataPtr)
				val = c.contextBlock.NewLoad(ifaceStruct)
			}
		}

		llvmArgs[i] = val
	}

	// Call function and return the result
	return value.Value{
		Value:      c.contextBlock.NewCall(fn.LlvmFunction, llvmArgs...),
		Type:       fn.ReturnType,
		IsVariable: false,
	}
}
