package compiler

import (
	"github.com/llir/llvm/ir"
	llvmTypes "github.com/llir/llvm/ir/types"
	"github.com/zegl/tre/compiler/compiler/types"
)

type ExternalFuncs struct {
	Printf  *types.Function
	Malloc  *types.Function
	Realloc *types.Function
	Memcpy  *types.Function
	Strcat  *types.Function
	Strcpy  *types.Function
	Strncpy *types.Function
	Strndup *types.Function
	Exit    *types.Function
}

func (c *Compiler) createExternalPackage() {
	externalPackageFuncs := make(map[string]*types.Function)

	{
		printfFunc := c.module.NewFunc("printf",
			i32.LLVM(),
			ir.NewParam("", llvmTypes.NewPointer(i8.LLVM())),
		)
		printfFunc.Sig.Variadic = true

		c.externalFuncs.Printf = &types.Function{
			LlvmReturnType: types.Void,
			FunctionName:   "printf",
			LlvmFunction:   printfFunc,
			IsExternal:     true,
		}
		externalPackageFuncs["Printf"] = c.externalFuncs.Printf
	}

	{
		c.externalFuncs.Malloc = &types.Function{
			LlvmReturnType: types.Void,
			FunctionName:   "malloc",
			LlvmFunction: c.module.NewFunc("malloc",
				llvmTypes.NewPointer(i8.LLVM()),
				ir.NewParam("", i64.LLVM()),
			),
			IsExternal: true,
		}
		externalPackageFuncs["malloc"] = c.externalFuncs.Malloc
	}

	{
		c.externalFuncs.Realloc = &types.Function{
			LlvmReturnType: types.Void,
			FunctionName:   "realloc",
			LlvmFunction: c.module.NewFunc("realloc",
				llvmTypes.NewPointer(i8.LLVM()),
				ir.NewParam("", llvmTypes.NewPointer(i8.LLVM())),
				ir.NewParam("", i64.LLVM()),
			),
			IsExternal: true,
		}
		externalPackageFuncs["realloc"] = c.externalFuncs.Realloc
	}

	{
		c.externalFuncs.Memcpy = &types.Function{
			LlvmReturnType: types.Void,
			FunctionName:   "memcpy",
			LlvmFunction: c.module.NewFunc("memcpy",
				llvmTypes.NewPointer(i8.LLVM()),
				ir.NewParam("dest", llvmTypes.NewPointer(i8.LLVM())),
				ir.NewParam("src", llvmTypes.NewPointer(i8.LLVM())),
				ir.NewParam("n", i64.LLVM()),
			),
			IsExternal: true,
		}
		externalPackageFuncs["memcpy"] = c.externalFuncs.Memcpy
	}

	{
		c.externalFuncs.Strcat = &types.Function{
			LlvmReturnType: types.Void,
			FunctionName:   "strcat",
			LlvmFunction: c.module.NewFunc("strcat",
				llvmTypes.NewPointer(i8.LLVM()),
				ir.NewParam("", llvmTypes.NewPointer(i8.LLVM())),
				ir.NewParam("", llvmTypes.NewPointer(i8.LLVM())),
			),
			IsExternal: true,
		}
		externalPackageFuncs["strcat"] = c.externalFuncs.Strcat
	}

	{
		c.externalFuncs.Strcpy = &types.Function{
			LlvmReturnType: types.Void,
			FunctionName:   "strcpy",
			LlvmFunction: c.module.NewFunc("strcpy",
				llvmTypes.NewPointer(i8.LLVM()),
				ir.NewParam("", llvmTypes.NewPointer(i8.LLVM())),
				ir.NewParam("", llvmTypes.NewPointer(i8.LLVM())),
			),
			IsExternal: true,
		}
		externalPackageFuncs["strcpy"] = c.externalFuncs.Strcpy
	}

	{
		c.externalFuncs.Strncpy = &types.Function{
			LlvmReturnType: types.Void,
			FunctionName:   "strncpy",
			LlvmFunction: c.module.NewFunc("strncpy",
				llvmTypes.NewPointer(i8.LLVM()),
				ir.NewParam("", llvmTypes.NewPointer(i8.LLVM())),
				ir.NewParam("", llvmTypes.NewPointer(i8.LLVM())),
				ir.NewParam("", i64.LLVM()),
			),
			IsExternal: true,
		}
		externalPackageFuncs["strncpy"] = c.externalFuncs.Strncpy
	}

	{
		c.externalFuncs.Strndup = &types.Function{
			LlvmReturnType: types.Void,
			FunctionName:   "strndup",
			LlvmFunction: c.module.NewFunc("strndup",
				llvmTypes.NewPointer(i8.LLVM()),
				ir.NewParam("", llvmTypes.NewPointer(i8.LLVM())),
				ir.NewParam("", i64.LLVM()),
			),
			IsExternal: true,
		}
		externalPackageFuncs["strndup"] = c.externalFuncs.Strndup
	}

	{
		c.externalFuncs.Exit = &types.Function{
			LlvmReturnType: types.Void,
			FunctionName:   "exit",
			LlvmFunction: c.module.NewFunc("exit",
				llvmTypes.Void,
				ir.NewParam("", i32.LLVM()),
			),
			IsExternal: true,
		}
		externalPackageFuncs["Exit"] = c.externalFuncs.Exit
	}

	c.packages["external"] = &types.PackageInstance{
		Funcs: externalPackageFuncs,
	}
}
