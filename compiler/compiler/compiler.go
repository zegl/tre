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
)

type Compiler struct {
	module *ir.Module

	// functions provided by the OS, such as printf and malloc
	externalFuncs ExternalFuncs

	// functions provided by the language, such as println
	globalFuncs map[string]*types.Function

	packages           map[string]*types.PackageInstance
	currentPackage     *types.PackageInstance
	currentPackageName string

	contextFunc *types.Function

	// Stack of return values pointers, is used both used if a function returns more
	// than one value (arg pointers), and single stack based returns
	contextFuncRetVals [][]value.Value

	contextBlock *ir.Block

	// Stack of variables that are in scope
	contextBlockVariables []map[string]value.Value

	// What a break or continue should resolve to
	contextLoopBreak    []*ir.Block
	contextLoopContinue []*ir.Block

	// Where a condition should jump when done
	contextCondAfter []*ir.Block

	// What type the current assign operation is assigning to.
	// Is used when evaluating what type an integer constant should be.
	contextAssignDest []value.Value

	// Stack of Alloc instructions
	// Is used to decide if values should be stack or heap allocated
	contextAlloc []*parser.AllocNode

	stringConstants map[string]*ir.Global

	// runtime.GOOS and runtime.GOARCH
	GOOS, GOARCH string
}

var (
	i8  = types.I8
	i32 = types.I32
	i64 = types.I64
)

func NewCompiler() *Compiler {
	c := &Compiler{
		module:      ir.NewModule(),
		globalFuncs: make(map[string]*types.Function),

		packages: make(map[string]*types.PackageInstance),

		contextFuncRetVals: make([][]value.Value, 0),

		contextBlockVariables: make([]map[string]value.Value, 0),

		contextLoopBreak:    make([]*ir.Block, 0),
		contextLoopContinue: make([]*ir.Block, 0),
		contextCondAfter:    make([]*ir.Block, 0),

		contextAssignDest: make([]value.Value, 0),

		stringConstants: make(map[string]*ir.Global),
	}

	c.createExternalPackage()
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

	// TODO: Allow cross compilation
	c.GOOS = runtime.GOOS
	c.GOARCH = runtime.GOARCH

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
	return c.module.String()
}

func (c *Compiler) addGlobal() {
	types.ModuleStringType = c.module.NewTypeDef("string", internal.String())

	// Create empty string constant
	types.EmptyStringConstant = c.module.NewGlobalDef(strings.NextStringName(), strings.Constant(""))
	types.EmptyStringConstant.Immutable = true

	// len_string
	strLen := internal.StringLen(types.ModuleStringType)
	c.globalFuncs["len_string"] = &types.Function{
		LlvmFunction:   strLen,
		LlvmReturnType: types.I64,
	}
	c.module.Funcs = append(c.module.Funcs, strLen)
}

func (c *Compiler) compile(instructions []parser.Node) {
	for _, i := range instructions {
		switch v := i.(type) {
		case *parser.ConditionNode:
			c.compileConditionNode(v)
		case *parser.DefineFuncNode:
			c.compileDefineFuncNode(v)
		case *parser.ReturnNode:
			c.compileReturnNode(v)
		case *parser.AllocNode:
			c.compileAllocNode(v)
		case *parser.AssignNode:
			c.compileAssignNode(v)
		case *parser.ForNode:
			c.compileForNode(v)
		case *parser.BreakNode:
			c.compileBreakNode(v)
		case *parser.ContinueNode:
			c.compileContinueNode(v)

		case *parser.DeclarePackageNode:
			// TODO: Make use of it
			break
		case *parser.ImportNode:
			// NOOP
			break

		case *parser.DefineTypeNode:
			t := parserTypeToType(v.Type)

			// Add type to module and override the structtype to use the named
			// type in the module
			if structType, ok := t.(*types.Struct); ok {
				structType.Type = c.module.NewTypeDef(v.Name, t.LLVM())
			}

			// Add to tre mapping
			typeConvertMap[v.Name] = t

		case *parser.SwitchNode:
			c.compileSwitchNode(v)

		default:
			c.compileValue(v)
			break
		}
	}
}

func (c *Compiler) funcByName(name string) (*types.Function, bool) {
	if f, ok := c.globalFuncs[name]; ok {
		return f, true
	}

	// Function in the current package
	if f, ok := c.currentPackage.Funcs[name]; ok {
		return f, true
	}

	return nil, false
}

func (c *Compiler) varByName(name string) value.Value {
	// Search scope in reverse (most specific first)
	for i := len(c.contextBlockVariables) - 1; i >= 0; i-- {
		if val, ok := c.contextBlockVariables[i][name]; ok {
			return val
		}
	}

	// Check functions in scope
	if fn, ok := c.funcByName(name); ok {
		return value.Value{Type: fn, Value: fn.LlvmFunction}
	}

	// Imported package?
	if pkg, ok := c.packages[name]; ok {
		return value.Value{
			Type: pkg,
		}
	}

	panic("undefined variable: " + name)
}

func (c *Compiler) setVar(name string, val value.Value) {
	c.contextBlockVariables[len(c.contextBlockVariables)-1][name] = val
}

func (c *Compiler) pushVariablesStack() {
	c.contextBlockVariables = append(c.contextBlockVariables, make(map[string]value.Value))
}

func (c *Compiler) popVariablesStack() {
	c.contextBlockVariables = c.contextBlockVariables[0 : len(c.contextBlockVariables)-1]
}

func (c *Compiler) compileValue(node parser.Node) value.Value {
	switch v := node.(type) {

	case *parser.ConstantNode:
		return c.compileConstantNode(v)
	case *parser.OperatorNode:
		return c.compileOperatorNode(v)
	case *parser.SubNode:
		return c.compileSubNode(v)
	case *parser.NameNode:
		return c.varByName(v.Name)
	case *parser.CallNode:
		return c.compileCallNode(v)
	case *parser.TypeCastNode:
		return c.compileTypeCastNode(v)
	case *parser.StructLoadElementNode:
		return c.compileStructLoadElementNode(v)
	case *parser.LoadArrayElement:
		return c.compileLoadArrayElement(v)
	case *parser.GetReferenceNode:
		return c.compileGetReferenceNode(v)
	case *parser.DereferenceNode:
		return c.compileDereferenceNode(v)
	case *parser.NegateNode:
		return c.compileNegateBoolNode(v)
	case *parser.InitializeSliceNode:
		return c.compileInitializeSliceNode(v)
	case *parser.SliceArrayNode:
		src := c.compileValue(v.Val)

		if _, ok := src.Type.(*types.StringType); ok {
			return c.compileSubstring(src, v)
		}

		return c.compileSliceArray(src, v)
	case *parser.InitializeStructNode:
		return c.compileInitStructWithValues(v)
	case *parser.TypeCastInterfaceNode:
		return c.compileTypeCastInterfaceNode(v)
	case *parser.DefineFuncNode:
		return c.compileDefineFuncNode(v)
	case *parser.InitializeArrayNode:
		return c.compileInitializeArrayNode(v)
	case *parser.DecrementNode:
		return c.compileDecrementNode(v)
	case *parser.IncrementNode:
		return c.compileIncrementNode(v)
	}

	panic("compileValue fail: " + fmt.Sprintf("%T: %+v", node, node))
}

func (c *Compiler) panic(block *ir.Block, message string) {
	globMsg := c.module.NewGlobalDef(strings.NextStringName(), strings.Constant("runtime panic: "+message+"\n"))
	globMsg.Immutable = true
	block.NewCall(c.externalFuncs.Printf.LlvmFunction, strings.Toi8Ptr(block, globMsg))
	block.NewCall(c.externalFuncs.Exit.LlvmFunction, constant.NewInt(llvmTypes.I32, 1))
}

type Panic string

func compilePanic(message string) {
	panic(Panic(fmt.Sprintf("compile panic: %s\n", message)))
}
