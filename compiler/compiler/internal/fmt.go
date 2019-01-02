package internal

import (
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/types"
	"github.com/zegl/tre/compiler/compiler/strings"
)

func Println(stringType types.Type, printfFunc *ir.Func, module *ir.Module) *ir.Func {
	param := ir.NewParam("input", stringType)
	res := ir.NewFunc("println", types.Void, param)
	block := res.NewBlock("entry")
	fmt := module.NewGlobalDef(strings.NextStringName(), strings.Constant("%s\n"))
	fmt.Immutable = true
	block.NewCall(printfFunc, strings.Toi8Ptr(block, fmt), strings.TreStringToi8Ptr(block, param))
	block.NewRet(nil)
	return res
}
