package compiler

import (
	"github.com/llir/llvm/ir"
	llvmTypes "github.com/llir/llvm/ir/types"

	"github.com/zegl/tre/compiler/compiler/types"
	"github.com/zegl/tre/compiler/compiler/value"
)

// ExternalFuncs and the "external" package contains a mapping to glibc functions.
// These are used to make bootstrapping of the language easier. The end goal is to not depend on glibc.
type ExternalFuncs struct {
	Printf  value.Value
	Malloc  value.Value
	Realloc value.Value
	Memcpy  value.Value
	Strcat  value.Value
	Strcpy  value.Value
	Strncpy value.Value
	Strndup value.Value
	Exit    value.Value
}

func (c *Compiler) createExternalPackage() {
	external := NewPkg("external")

	setExternal := func(internalName string, fn *ir.Func, variadic bool) value.Value {
		fn.Sig.Variadic = variadic
		val := value.Value{
			Type: &types.Function{
				LlvmReturnType: types.Void,
				FuncType:       fn.Type(),
				IsExternal:     true,
			},
			Value: fn,
		}
		external.DefinePkgVar(internalName, val)
		return val
	}

	c.externalFuncs.Printf = setExternal("Printf", c.module.NewFunc("printf",
		i32.LLVM(),
		ir.NewParam("", llvmTypes.NewPointer(i8.LLVM())),
	), true)

	c.externalFuncs.Malloc = setExternal("malloc", c.module.NewFunc("malloc",
		llvmTypes.NewPointer(i8.LLVM()),
		ir.NewParam("", i64.LLVM()),
	), false)

	c.externalFuncs.Realloc = setExternal("realloc", c.module.NewFunc("realloc",
		llvmTypes.NewPointer(i8.LLVM()),
		ir.NewParam("", llvmTypes.NewPointer(i8.LLVM())),
		ir.NewParam("", i64.LLVM()),
	), false)

	c.externalFuncs.Memcpy = setExternal("memcpy", c.module.NewFunc("memcpy",
		llvmTypes.NewPointer(i8.LLVM()),
		ir.NewParam("dest", llvmTypes.NewPointer(i8.LLVM())),
		ir.NewParam("src", llvmTypes.NewPointer(i8.LLVM())),
		ir.NewParam("n", i64.LLVM()),
	), false)

	c.externalFuncs.Strcat = setExternal("strcat", c.module.NewFunc("strcat",
		llvmTypes.NewPointer(i8.LLVM()),
		ir.NewParam("", llvmTypes.NewPointer(i8.LLVM())),
		ir.NewParam("", llvmTypes.NewPointer(i8.LLVM())),
	), false)

	c.externalFuncs.Strcpy = setExternal("strcpy", c.module.NewFunc("strcpy",
		llvmTypes.NewPointer(i8.LLVM()),
		ir.NewParam("", llvmTypes.NewPointer(i8.LLVM())),
		ir.NewParam("", llvmTypes.NewPointer(i8.LLVM())),
	), false)

	c.externalFuncs.Strncpy = setExternal("strncpy", c.module.NewFunc("strncpy",
		llvmTypes.NewPointer(i8.LLVM()),
		ir.NewParam("", llvmTypes.NewPointer(i8.LLVM())),
		ir.NewParam("", llvmTypes.NewPointer(i8.LLVM())),
		ir.NewParam("", i64.LLVM()),
	), false)

	c.externalFuncs.Strndup = setExternal("strndup", c.module.NewFunc("strndup",
		llvmTypes.NewPointer(i8.LLVM()),
		ir.NewParam("", llvmTypes.NewPointer(i8.LLVM())),
		ir.NewParam("", i64.LLVM()),
	), false)

	c.externalFuncs.Exit = setExternal("exit", c.module.NewFunc("exit",
		llvmTypes.Void,
		ir.NewParam("", i32.LLVM()),
	), false)

	c.packages["external"] = external
}
