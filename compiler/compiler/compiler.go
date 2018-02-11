package compiler

import (
	"fmt"
	"os"
	"runtime"

	"github.com/zegl/tre/compiler/compiler/internal"
	"github.com/zegl/tre/compiler/compiler/strings"
	"github.com/zegl/tre/compiler/parser"

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
		globalFuncs:   make(map[string]*ir.Function),

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
	printfFunc := c.module.NewFunction("printf", i32, ir.NewParam("", types.NewPointer(i8)))
	printfFunc.Sig.Variadic = true
	c.externalFuncs["printf"] = printfFunc

	c.externalFuncs["strcat"] = c.module.NewFunction("strcat",
		types.NewPointer(i8),
		ir.NewParam("", types.NewPointer(i8)),
		ir.NewParam("", types.NewPointer(i8)))

	c.externalFuncs["strcpy"] = c.module.NewFunction("strcpy",
		types.NewPointer(i8),
		ir.NewParam("", types.NewPointer(i8)),
		ir.NewParam("", types.NewPointer(i8)))

	c.externalFuncs["strncpy"] = c.module.NewFunction("strncpy",
		types.NewPointer(i8),
		ir.NewParam("", types.NewPointer(i8)),
		ir.NewParam("", types.NewPointer(i8)),
		ir.NewParam("", i64),
	)

	c.externalFuncs["strndup"] = c.module.NewFunction("strndup",
		types.NewPointer(i8),
		ir.NewParam("", types.NewPointer(i8)),
		ir.NewParam("", i64),
	)

	c.externalFuncs["exit"] = c.module.NewFunction("exit",
		types.Void,
		ir.NewParam("", i32),
	)
}

func (c *compiler) addGlobal() {
	typeConvertMap["string"] = c.module.NewType("string", internal.String())

	// printf := internal.Printf(typeConvertMap["string"], c.externalFuncs["printf"])
	// c.module.AppendFunction(printf)
	c.globalFuncs["printf"] = c.externalFuncs["printf"]

	c.globalFuncs["println"] = internal.Println(
		typeConvertMap["string"],
		c.externalFuncs["printf"],
		c.module,
	)
	c.module.AppendFunction(c.globalFuncs["println"])

	// String function
	strLen := internal.StringLen(typeConvertMap["string"])
	c.module.AppendFunction(strLen)
	c.globalFuncs["len_string"] = strLen
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
			cond := c.compileCondition(v.Cond)

			afterBlock := function.NewBlock(getBlockName() + "-after")
			trueBlock := function.NewBlock(getBlockName() + "-true")
			falseBlock := afterBlock

			if len(v.False) > 0 {
				falseBlock = function.NewBlock(getBlockName() + "-false")
			}

			block.NewCondBr(cond, trueBlock, falseBlock)

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
			params := make([]*types.Param, len(v.Arguments))
			for k, par := range v.Arguments {
				params[k] = ir.NewParam(par.Name, typeStringToLLVM(par.Type))
			}

			funcRetType := types.Type(types.Void)
			if len(v.ReturnValues) == 1 {
				funcRetType = typeStringToLLVM(v.ReturnValues[0].Type)
			}

			// TODO: Only do this in the main package
			if v.Name == "main" {
				if len(v.ReturnValues) != 0 {
					panic("main func can not have a return type")
				}
				funcRetType = typeStringToLLVM("int32")
			}

			// Create a new function, and add it to the list of global functions
			fn := c.module.NewFunction(v.Name, funcRetType, params...)
			c.globalFuncs[v.Name] = fn

			entry := fn.NewBlock(getBlockName())

			c.contextFunc = fn
			c.contextBlock = entry
			c.contextBlockVariables = make(map[string]value.Value)

			// Save all parameters in the block mapping
			for i, param := range params {
				paramName := v.Arguments[i].Name

				// Structs needs to be pointer-allocated
				if _, ok := param.Type().(*types.StructType); ok {
					paramPtr := entry.NewAlloca(typeStringToLLVM(v.Arguments[i].Type))
					entry.NewStore(param, paramPtr)
					c.contextBlockVariables[paramName] = paramPtr
					continue
				}

				c.contextBlockVariables[paramName] = param
			}

			c.compile(v.Body)

			// Return void if there is no return type explicitly set
			if len(v.ReturnValues) == 0 {
				c.contextBlock.NewRet(nil)
			}

			// Return 0 by default in main func
			if v.Name == "main" {
				c.contextBlock.NewRet(constant.NewInt(0, typeStringToLLVM("int32")))
			}

			break

		case parser.ReturnNode:
			// Set value and jump to return block
			val := c.compileValue(v.Val)

			if loadNeeded(val) {
				val = block.NewLoad(val)
			}

			block.NewRet(val)
			break

		case parser.AllocNode:

			// Allocate from type
			if typeNode, ok := v.Val.(parser.TypeNode); ok {
				alloc := block.NewAlloca(typeNodeToLLVMType(typeNode))
				alloc.SetName(v.Name)
				c.contextBlockVariables[v.Name] = alloc
				break
			}

			// Allocate from value
			val := c.compileValue(v.Val)

			if loadNeeded(val) {
				val = block.NewLoad(val)
			}

			alloc := block.NewAlloca(val.Type())
			alloc.SetName(v.Name)
			block.NewStore(val, alloc)
			c.contextBlockVariables[v.Name] = alloc

			break

		case parser.AssignNode:
			var dst value.Value

			// TODO: Remove AssignNode.Name
			if len(v.Name) > 0 {
				dst = c.varByName(v.Name)
			} else {
				dst = c.compileValue(v.Target)
			}

			// Allocate from type
			if typeNode, ok := v.Val.(parser.TypeNode); ok {
				if singleTypeNode, ok := typeNode.(*parser.SingleTypeNode); ok {
					alloc := block.NewAlloca(typeStringToLLVM(singleTypeNode.TypeName))
					block.NewStore(block.NewLoad(alloc), dst)
					break
				}

				panic("AssignNode from non TypeNode is not allowed")
			}

			// Allocate from value
			val := c.compileValue(v.Val)

			if loadNeeded(val) {
				val = block.NewLoad(val)
			}

			block.NewStore(val, dst)
			break

		case parser.DefineTypeNode:

			if structNode, ok := v.Type.(*parser.StructTypeNode); ok {

				var structTypes []types.Type
				for _, t := range structNode.Types {
					if singleTypeNode, ok := t.(*parser.SingleTypeNode); ok {
						structTypes = append(structTypes, typeStringToLLVM(singleTypeNode.TypeName))
					} else {
						panic("unable to define node Type. nested structs are not supported")
					}
				}

				structType := types.NewStruct(structTypes...)

				// Add to tre mapping
				typeConvertMap[v.Name] = structType
				typeMapElementNameIndex[v.Name] = structNode.Names

				// Generate LLVM code
				c.module.NewType(v.Name, structType)
				break
			}

			panic("unable to define node Type")

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
	switch v := node.(type) {

	case parser.ConstantNode:
		switch v.Type {
		case parser.NUMBER:
			return constant.NewInt(v.Value, i64)

		case parser.STRING:
			constString := c.module.NewGlobalDef(strings.NextStringName(), strings.Constant(v.ValueStr))
			constString.IsConst = true

			alloc := c.contextBlock.NewAlloca(typeConvertMap["string"])

			// Save length of the string
			lenItem := c.contextBlock.NewGetElementPtr(alloc, constant.NewInt(0, i32), constant.NewInt(0, i32))
			c.contextBlock.NewStore(constant.NewInt(int64(len(v.ValueStr)), i64), lenItem)

			// Save i8* version of string
			strItem := c.contextBlock.NewGetElementPtr(alloc, constant.NewInt(0, i32), constant.NewInt(1, i32))
			c.contextBlock.NewStore(strings.Toi8Ptr(c.contextBlock, constString), strItem)
			return c.contextBlock.NewLoad(alloc)
		}
		break

	case parser.OperatorNode:
		left := c.compileValue(v.Left)
		right := c.compileValue(v.Right)

		if loadNeeded(left) {
			left = c.contextBlock.NewLoad(left)
		}

		if loadNeeded(right) {
			right = c.contextBlock.NewLoad(right)
		}

		if !left.Type().Equal(right.Type()) {
			panic(fmt.Sprintf("Different types in operation: %T and %T", left, right))
		}

		switch left.Type().GetName() {
		case "string":
			if v.Operator == parser.OP_ADD {
				leftLen := c.contextBlock.NewExtractValue(left, []int64{0})
				rightLen := c.contextBlock.NewExtractValue(right, []int64{0})
				sumLen := c.contextBlock.NewAdd(leftLen, rightLen)

				backingArray := c.contextBlock.NewAlloca(i8)
				backingArray.NElems = sumLen

				// Copy left to new backing array
				c.contextBlock.NewCall(c.externalFuncs["strcpy"], backingArray, c.contextBlock.NewExtractValue(left, []int64{1}))

				// Append right to backing array
				c.contextBlock.NewCall(c.externalFuncs["strcat"], backingArray, c.contextBlock.NewExtractValue(right, []int64{1}))

				alloc := c.contextBlock.NewAlloca(typeConvertMap["string"])

				// Save length of the string
				lenItem := c.contextBlock.NewGetElementPtr(alloc, constant.NewInt(0, i32), constant.NewInt(0, i32))
				c.contextBlock.NewStore(sumLen, lenItem)

				// Save i8* version of string
				strItem := c.contextBlock.NewGetElementPtr(alloc, constant.NewInt(0, i32), constant.NewInt(1, i32))
				c.contextBlock.NewStore(backingArray, strItem)
				return c.contextBlock.NewLoad(alloc)
			}

			panic("string does not implement operation " + v.Operator)
		}

		switch v.Operator {
		case parser.OP_ADD:
			return c.contextBlock.NewAdd(left, right)
		case parser.OP_SUB:
			return c.contextBlock.NewSub(left, right)
		case parser.OP_MUL:
			return c.contextBlock.NewMul(left, right)
		case parser.OP_DIV:
			return c.contextBlock.NewSDiv(left, right) // SDiv == Signed Division
		}
		break

	case parser.NameNode:
		return c.varByName(v.Name)

	case parser.CallNode:
		var args []value.Value

		// len() functions
		if v.Function == "len" {
			arg := c.compileValue(v.Arguments[0])
			if arg.Type().String() == "%string*" || arg.Type().String() == "%string" {
				if arg.Type().String() == "%string*" {
					arg = c.contextBlock.NewLoad(arg)
				}
				return c.contextBlock.NewCall(c.funcByName("len_string"), arg)
			}

			if ptrType, ok := arg.Type().(*types.PointerType); ok {
				if arrayType, ok := ptrType.Elem.(*types.ArrayType); ok {
					return constant.NewInt(arrayType.Len, i64)
				}
			}
		}

		_, isExternal := c.externalFuncs[v.Function]

		for _, vv := range v.Arguments {
			val := c.compileValue(vv)

			// Convert %string* to i8* when calling external functions
			if isExternal {
				if val.Type().String() == "%string*" {
					val = c.contextBlock.NewLoad(val)
				}
				if val.Type().String() == "%string" {
					args = append(args, c.contextBlock.NewExtractValue(val, []int64{1}))
					continue
				}
			}

			if loadNeeded(val) {
				val = c.contextBlock.NewLoad(val)
			}

			args = append(args, val)
		}

		fn := c.funcByName(v.Function)

		// Call function and return the result
		return c.contextBlock.NewCall(fn, args...)

	case parser.TypeCastNode:
		val := c.compileValue(v.Val)

		var current *types.IntType
		var ok bool

		if loadNeeded(val) {
			currentType := ptrTypeType(val)
			current, ok = currentType.(*types.IntType)
			if !ok {
				panic("TypeCast origin must be int type")
			}
		} else {
			current, ok = val.Type().(*types.IntType)
			if !ok {
				panic("TypeCast origin must be int type")
			}
		}

		target, ok := typeStringToLLVM(v.Type).(*types.IntType)
		if !ok {
			panic("TypeCast target must be int type")
		}

		if loadNeeded(val) {
			val = c.contextBlock.NewLoad(val)
		}

		// Same size, nothing to do here
		if current.Size == target.Size {
			return val
		}

		res := c.contextBlock.NewAlloca(target)

		var changedSize value.Value

		if current.Size < target.Size {
			changedSize = c.contextBlock.NewSExt(val, target)
		} else {
			changedSize = c.contextBlock.NewTrunc(val, target)
		}

		c.contextBlock.NewStore(changedSize, res)
		return res

	case parser.StructLoadElementNode:
		src := c.compileValue(v.Struct)

		typeName := src.Type().String()

		if ptrType, ok := src.Type().(*types.PointerType); ok {
			// Get type behind pointer
			typeName = ptrType.Elem.String()
		} else {
			// GetElementPtr only works on pointer types, and we don't have a pointer to our object. Allocate it and
			// use the pointer instead
			dst := c.contextBlock.NewAlloca(src.Type())
			c.contextBlock.NewStore(src, dst)
			src = dst
		}

		// Remove % from the name
		typeName = typeName[1:]

		indexMapping, ok := typeMapElementNameIndex[typeName]
		if !ok {
			panic(fmt.Sprintf("%s internal error: no such type map indexing", typeName))
		}

		elementIndex, ok := indexMapping[v.ElementName]
		if !ok {
			panic(fmt.Sprintf("%s has no such element: %s", src.Type(), v.ElementName))
		}

		return c.contextBlock.NewGetElementPtr(src, constant.NewInt(0, i32), constant.NewInt(int64(elementIndex), i32))

	case parser.SliceArrayNode:
		src := c.compileValue(v.Val)

		var originalLength *ir.InstExtractValue

		// Get backing array from string type
		if src.Type().String() == "%string*" {
			src = c.contextBlock.NewLoad(src)
		}
		if src.Type().String() == "%string" {
			originalLength = c.contextBlock.NewExtractValue(src, []int64{0})
			src = c.contextBlock.NewExtractValue(src, []int64{1})
		}

		start := c.compileValue(v.Start)

		outsideOfLengthBr := c.contextBlock.Parent.NewBlock(getBlockName())
		c.panic(outsideOfLengthBr, "Substring start larger than len")
		outsideOfLengthBr.NewUnreachable()

		safeBlock := c.contextBlock.Parent.NewBlock(getBlockName())

		// Make sure that the offset is within the string length
		cmp := c.contextBlock.NewICmp(ir.IntUGE, start, originalLength)
		c.contextBlock.NewCondBr(cmp, outsideOfLengthBr, safeBlock)

		c.contextBlock = safeBlock

		offset := safeBlock.NewGetElementPtr(src, start)

		var length value.Value
		if v.HasEnd {
			length = c.compileValue(v.End)
		} else {
			length = constant.NewInt(1, i64)
		}

		dst := safeBlock.NewCall(c.externalFuncs["strndup"], offset, length)

		// Convert *i8 to %string
		alloc := safeBlock.NewAlloca(typeConvertMap["string"])

		// Save length of the string
		lenItem := safeBlock.NewGetElementPtr(alloc, constant.NewInt(0, i32), constant.NewInt(0, i32))
		safeBlock.NewStore(constant.NewInt(100, i64), lenItem) // TODO

		// Save i8* version of string
		strItem := safeBlock.NewGetElementPtr(alloc, constant.NewInt(0, i32), constant.NewInt(1, i32))
		safeBlock.NewStore(dst, strItem)
		return safeBlock.NewLoad(alloc)

	case parser.LoadArrayElement:
		arr := c.compileValue(v.Array)
		index := c.compileValue(v.Pos)

		var runtimeLength value.Value
		var compileTimeLenght int64
		lengthKnownAtCompileTime := false
		lengthKnownAtRunTime := false
		arrIsString := false

		// Length of array
		if ptrType, ok := arr.Type().(*types.PointerType); ok {
			if arrayType, ok := ptrType.Elem.(*types.ArrayType); ok {
				compileTimeLenght = arrayType.Len
				lengthKnownAtCompileTime = true
			}
		}

		// Length of string
		if !lengthKnownAtCompileTime {
			// Get backing array from string type
			if arr.Type().String() == "%string*" {
				arr = c.contextBlock.NewLoad(arr)
			}
			if arr.Type().String() == "%string" {
				runtimeLength = c.contextBlock.NewExtractValue(arr, []int64{0})
				// Get backing array
				arr = c.contextBlock.NewExtractValue(arr, []int64{1})
				lengthKnownAtRunTime = true
				arrIsString = true
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

			if intType, ok := index.(*constant.Int); ok {
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
				c.contextBlock.NewICmp(ir.IntSLT, index, constant.NewInt(0, i64)),
				c.contextBlock.NewICmp(ir.IntSGE, index, runtimeLength),
			)

			c.contextBlock.NewCondBr(outOfRangeCmp, outsideOfLengthBlock, safeBlock)

			c.contextBlock = safeBlock
		}

		var indicies []value.Value
		if !arrIsString {
			indicies = append(indicies, constant.NewInt(0, i64))
		}
		indicies = append(indicies, index)

		return c.contextBlock.NewGetElementPtr(arr, indicies...)
	}

	panic("compileValue fail: " + fmt.Sprintf("%T: %+v", node, node))
}

func loadNeeded(val value.Value) bool {
	t := val.Type()
	if _, ok := t.(*types.PointerType); ok {
		return true
	}
	return false
}

func (c *compiler) panic(block *ir.BasicBlock, message string) {
	globMsg := c.module.NewGlobalDef(strings.NextStringName(), strings.Constant("runtime panic: "+message+"\n"))
	globMsg.IsConst = true
	block.NewCall(c.externalFuncs["printf"], strings.Toi8Ptr(block, globMsg))
	block.NewCall(c.externalFuncs["exit"], constant.NewInt(1, i32))
}

func compilePanic(message string) {
	fmt.Fprintf(os.Stderr, "compile panic: %s\n", message)
	os.Exit(1)
}
