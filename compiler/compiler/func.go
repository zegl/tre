package compiler

import (
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	llvmTypes "github.com/llir/llvm/ir/types"
	llvmValue "github.com/llir/llvm/ir/value"
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
}

func (c *Compiler) compileReturnNode(v parser.ReturnNode) {
	// Set value and jump to return block
	val := c.compileValue(v.Val)

	if val.PointerLevel > 0 {
		c.contextBlock.NewRet(c.contextBlock.NewLoad(val.Value))
		return
	}

	c.contextBlock.NewRet(val.Value)
}

func (c *Compiler) compileCallNode(v parser.CallNode) value.Value {
	var args []llvmValue.Value

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
}
