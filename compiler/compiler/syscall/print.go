package syscall

import (
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	llvmTypes "github.com/llir/llvm/ir/types"

	"github.com/zegl/tre/compiler/compiler/strings"
	"github.com/zegl/tre/compiler/compiler/types"
	"github.com/zegl/tre/compiler/compiler/value"
)

func Print(block *ir.Block, value value.Value, goos string) {
	asmFunc := ir.NewInlineAsm(llvmTypes.NewPointer(llvmTypes.NewFunc(types.I64.LLVM())), "syscall", "=r,{rax},{rdi},{rsi},{rdx}")
	asmFunc.SideEffect = true

	strPtr := strings.TreToI8Ptr(block, value.Value)
	strLen := strings.Len(block, value.Value)

	block.NewCall(asmFunc,
		constant.NewInt(types.I64.Type, Convert(WRITE, goos)), // rax
		constant.NewInt(types.I64.Type, 1),                    // rdi, stdout
		strPtr,                                                // rsi
		strLen,                                                // rdx
	)
}
