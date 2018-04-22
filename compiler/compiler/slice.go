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
	arrType := src.Type.(*types.Array)

	sliceType := internal.Slice(arrType.Type.LLVM())

	alloc := c.contextBlock.NewAlloca(sliceType)

	startIndex := c.compileValue(v.Start)
	endIndex := c.compileValue(v.End)

	sliceLen := c.contextBlock.NewSub(endIndex.Value, startIndex.Value)
	sliceLen32 := c.contextBlock.NewTrunc(sliceLen, i32.LLVM())

	offset32 := c.contextBlock.NewTrunc(startIndex.Value, i32.LLVM())

	// Len
	lenItem := c.contextBlock.NewGetElementPtr(alloc, constant.NewInt(0, i32.LLVM()), constant.NewInt(0, i32.LLVM()))
	c.contextBlock.NewStore(sliceLen32, lenItem)

	// Cap
	capItem := c.contextBlock.NewGetElementPtr(alloc, constant.NewInt(0, i32.LLVM()), constant.NewInt(1, i32.LLVM()))
	c.contextBlock.NewStore(sliceLen32, capItem)

	// Offset
	offsetItem := c.contextBlock.NewGetElementPtr(alloc, constant.NewInt(0, i32.LLVM()), constant.NewInt(2, i32.LLVM()))
	c.contextBlock.NewStore(offset32, offsetItem)

	// Backing Array
	backingArrayItem := c.contextBlock.NewGetElementPtr(alloc, constant.NewInt(0, i32.LLVM()), constant.NewInt(3, i32.LLVM()))
	_ = backingArrayItem

	itemPtr := c.contextBlock.NewBitCast(src.Value, llvmTypes.NewPointer(arrType.Type.LLVM()))
	c.contextBlock.NewStore(itemPtr, backingArrayItem)

	res := value.Value{
		Type: &types.Slice{
			Type:     arrType.Type,
			LlvmType: sliceType,
		},
		Value: alloc,
	}

	return res
}

func (c *Compiler) appendFuncCall(v parser.CallNode) value.Value {
	// 1. Grow the backing array if neccesary (cap == len)
	// 1.1. Create a new array (at least double the size).
	// 1.2. Copy all data
	// 1.3. Reset the offset
	// 3. Increase len by 1
	// 4. Return the new slice

	input := c.compileValue(v.Arguments[0])
	inputSlice := input.Type.(*types.Slice)

	preAppendContextBlock := c.contextBlock

	len := preAppendContextBlock.NewGetElementPtr(input.Value, constant.NewInt(0, i32.LLVM()), constant.NewInt(0, i32.LLVM()))
	len.SetName(getVarName("len"))

	cap := preAppendContextBlock.NewGetElementPtr(input.Value, constant.NewInt(0, i32.LLVM()), constant.NewInt(1, i32.LLVM()))
	cap.SetName(getVarName("cap"))

	// Create blocks that are needed lager
	growSliceAllocBlock := c.contextFunc.NewBlock(getBlockName() + "-grow-slice-alloc")
	appendToSliceBlock := c.contextFunc.NewBlock(getBlockName() + "-append-to-slice")

	loadedLen := preAppendContextBlock.NewLoad(len)
	loadedCap := preAppendContextBlock.NewLoad(cap)

	// Grow backing array if len == cap
	preAppendContextBlock.NewCondBr(
		c.contextBlock.NewICmp(ir.IntEQ, loadedLen, loadedCap),
		growSliceAllocBlock,
		appendToSliceBlock,
	)

	len64 := growSliceAllocBlock.NewSExt(loadedLen, i64.LLVM())

	// Double the size
	newCap := growSliceAllocBlock.NewAlloca(i64.LLVM())
	newCap.SetName(getVarName("new-cap"))
	newCapMul := growSliceAllocBlock.NewMul(len64, constant.NewInt(2, i64.LLVM()))
	growSliceAllocBlock.NewStore(newCapMul, newCap)

	reallocSize := growSliceAllocBlock.NewMul(newCapMul, constant.NewInt(inputSlice.Type.Size(), i64.LLVM()))

	prevBackingArrayPtr := growSliceAllocBlock.NewGetElementPtr(input.Value, constant.NewInt(0, i32.LLVM()), constant.NewInt(3, i32.LLVM()))
	prevBackingArrayPtr.SetName(getVarName("raw-prev-backing-ptr"))

	loadedPrevBacking := growSliceAllocBlock.NewLoad(prevBackingArrayPtr)

	castedPrevBacking := growSliceAllocBlock.NewBitCast(loadedPrevBacking, llvmTypes.NewPointer(llvmTypes.I8))
	castedPrevBacking.SetName(getVarName("casted-prev-backing-ptr"))

	reallocatedSpaceRaw := growSliceAllocBlock.NewCall(c.externalFuncs["realloc"], castedPrevBacking, reallocSize)
	reallocatedSpaceRaw.SetName(getVarName("reallocatedspace-raw"))

	reallocatedSpace := growSliceAllocBlock.NewBitCast(reallocatedSpaceRaw, llvmTypes.NewPointer(inputSlice.Type.LLVM()))
	reallocatedSpace.SetName(getVarName("reallocatedspace-casted"))

	// Save cap
	sliceCapPtr := growSliceAllocBlock.NewGetElementPtr(input.Value, constant.NewInt(0, i32.LLVM()), constant.NewInt(1, i32.LLVM()))
	newCap32 := growSliceAllocBlock.NewTrunc(growSliceAllocBlock.NewLoad(newCap), i32.LLVM())
	growSliceAllocBlock.NewStore(newCap32, sliceCapPtr)

	// Set offset to 0
	sliceOffsetPtr := growSliceAllocBlock.NewGetElementPtr(input.Value, constant.NewInt(0, i32.LLVM()), constant.NewInt(2, i32.LLVM()))
	growSliceAllocBlock.NewStore(constant.NewInt(0, i32.LLVM()), sliceOffsetPtr)

	// Set as new backing array
	backingArrayPtr := growSliceAllocBlock.NewGetElementPtr(input.Value, constant.NewInt(0, i32.LLVM()), constant.NewInt(3, i32.LLVM()))
	growSliceAllocBlock.NewStore(reallocatedSpace, backingArrayPtr)

	growSliceAllocBlock.NewBr(appendToSliceBlock)

	// Add item

	// Get current len
	sliceLenPtr := appendToSliceBlock.NewGetElementPtr(input.Value, constant.NewInt(0, i32.LLVM()), constant.NewInt(0, i32.LLVM()))
	sliceLen := appendToSliceBlock.NewLoad(sliceLenPtr)
	sliceLen.SetName(getVarName("slicelen"))

	// TODO: Allow adding many items at once `foo = append(foo, bar, baz)`
	backingArrayAppendPtr := appendToSliceBlock.NewGetElementPtr(input.Value,
		constant.NewInt(0, i32.LLVM()),
		constant.NewInt(3, i32.LLVM()),
	)

	backingArrayAppendPtr.SetName(getVarName("backingarrayptr"))
	loadedPtr := appendToSliceBlock.NewLoad(backingArrayAppendPtr)
	storePtr := appendToSliceBlock.NewGetElementPtr(loadedPtr, sliceLen)

	c.contextBlock = appendToSliceBlock
	newVal := c.compileValue(v.Arguments[1]).Value
	appendToSliceBlock.NewStore(newVal, storePtr)

	// Increase len

	newLen := appendToSliceBlock.NewAdd(sliceLen, constant.NewInt(1, i32.LLVM()))
	appendToSliceBlock.NewStore(newLen, sliceLenPtr)

	c.contextBlock = appendToSliceBlock

	return input
}
