package compiler

import (
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	llvmValue "github.com/llir/llvm/ir/value"
	"github.com/zegl/tre/compiler/compiler/value"
	"github.com/zegl/tre/compiler/parser"
)

func (c *Compiler) compileSubstring(src value.Value, v parser.SliceArrayNode) value.Value {
	srcVal := src.Value

	var originalLength *ir.InstExtractValue

	// Get backing array from string type
	if src.PointerLevel > 0 {
		srcVal = c.contextBlock.NewLoad(srcVal)
	}
	if src.Type.Name() == "string" {
		originalLength = c.contextBlock.NewExtractValue(srcVal, []int64{0})
		srcVal = c.contextBlock.NewExtractValue(srcVal, []int64{1})
	}

	start := c.compileValue(v.Start)

	outsideOfLengthBr := c.contextBlock.Parent.NewBlock(getBlockName())
	c.panic(outsideOfLengthBr, "Substring start larger than len")
	outsideOfLengthBr.NewUnreachable()

	safeBlock := c.contextBlock.Parent.NewBlock(getBlockName())

	// Make sure that the offset is within the string length
	cmp := c.contextBlock.NewICmp(ir.IntUGE, start.Value, originalLength)
	c.contextBlock.NewCondBr(cmp, outsideOfLengthBr, safeBlock)

	c.contextBlock = safeBlock

	offset := safeBlock.NewGetElementPtr(srcVal, start.Value)

	var length llvmValue.Value
	if v.HasEnd {
		length = c.compileValue(v.End).Value
	} else {
		length = constant.NewInt(1, i64.LLVM())
	}

	dst := safeBlock.NewCall(c.externalFuncs["strndup"], offset, length)

	// Convert *i8 to %string
	alloc := safeBlock.NewAlloca(typeConvertMap["string"].LLVM())

	// Save length of the string
	lenItem := safeBlock.NewGetElementPtr(alloc, constant.NewInt(0, i32.LLVM()), constant.NewInt(0, i32.LLVM()))
	safeBlock.NewStore(constant.NewInt(100, i64.LLVM()), lenItem) // TODO

	// Save i8* version of string
	strItem := safeBlock.NewGetElementPtr(alloc, constant.NewInt(0, i32.LLVM()), constant.NewInt(1, i32.LLVM()))
	safeBlock.NewStore(dst, strItem)

	return value.Value{
		Value:        safeBlock.NewLoad(alloc),
		Type:         typeConvertMap["string"],
		PointerLevel: 0,
	}
}

func (c *Compiler) compileSliceArray(src value.Value, v parser.SliceArrayNode) value.Value {
	// TODO
	return c.compileSubstring(src, v)
}
