package compiler

import (
	"github.com/zegl/tre/compiler/compiler/syscall"
	"github.com/zegl/tre/compiler/compiler/types"
	"github.com/zegl/tre/compiler/compiler/value"
	"github.com/zegl/tre/compiler/parser"
)

func (c *Compiler) printFuncCall(v *parser.CallNode) value.Value {
	arg := c.compileValue(v.Arguments[0])
	syscall.Print(c.contextBlock, arg, c.GOOS)
	return value.Value{Type: types.Void}
}
