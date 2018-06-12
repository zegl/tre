package compiler

import (
	"fmt"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	llvmTypes "github.com/llir/llvm/ir/types"
	llvmValue "github.com/llir/llvm/ir/value"
	"github.com/zegl/tre/compiler/compiler/internal"
	"github.com/zegl/tre/compiler/compiler/types"
	"github.com/zegl/tre/compiler/compiler/value"
	"github.com/zegl/tre/compiler/parser"
)

func (c *Compiler) compileSubstring(src value.Value, v *parser.SliceArrayNode) value.Value {
	srcVal := src.Value

	var originalLength *ir.InstExtractValue

	// Get backing array from string type
	if src.IsVariable {
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

	dst := safeBlock.NewCall(c.externalFuncs.Strndup.LlvmFunction, offset, length)

	// Convert *i8 to %string
	alloc := safeBlock.NewAlloca(typeConvertMap["string"].LLVM())

	// Save length of the string
	lenItem := safeBlock.NewGetElementPtr(alloc, constant.NewInt(0, i32.LLVM()), constant.NewInt(0, i32.LLVM()))
	safeBlock.NewStore(constant.NewInt(100, i64.LLVM()), lenItem) // TODO

	// Save i8* version of string
	strItem := safeBlock.NewGetElementPtr(alloc, constant.NewInt(0, i32.LLVM()), constant.NewInt(1, i32.LLVM()))
	safeBlock.NewStore(dst, strItem)

	return value.Value{
		Value:      safeBlock.NewLoad(alloc),
		Type:       typeConvertMap["string"],
		IsVariable: false,
	}
}

func (c *Compiler) compileSliceArray(src value.Value, v *parser.SliceArrayNode) value.Value {
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

func (c *Compiler) appendFuncCall(v *parser.CallNode) value.Value {
	// 1. Grow the backing array if neccesary (cap == len)
	// 1.1. Create a new array (at least double the size).
	// 1.2. Copy all data
	// 1.3. Reset the offset
	// 3. Increase len by 1
	// 4. Return the new slice

	input := c.compileValue(v.Arguments[0])
	inputSlice := input.Type.(*types.Slice)

	isSelfAssign := false

	// Check if this the slice is currently assigning to itself.
	// If that is the case (which it commonly is), we can safely expand the backing array.
	// If not: The whole slice + backing array has to be copied before it can be altered.
	if len(c.contextAssignDest) > 0 {
		assignDst := c.contextAssignDest[len(c.contextAssignDest)-1]
		if assignDst.Value.Ident() == input.Value.Ident() {
			isSelfAssign = true
		}
	}

	// Will contain the slice that the new data will be added to
	// Can be the same slice as "input", or a new one
	var sliceToAppendTo value.Value

	// Create blocks that are needed lager
	appendToSliceBlock := c.contextBlock.Parent.NewBlock(getBlockName() + "-append-to-slice")

	if isSelfAssign {
		growSliceAllocBlock := c.contextBlock.Parent.NewBlock(getBlockName() + "-grow-slice-alloc")

		preAppendContextBlock := c.contextBlock

		len := preAppendContextBlock.NewGetElementPtr(input.Value, constant.NewInt(0, i32.LLVM()), constant.NewInt(0, i32.LLVM()))
		len.SetName(getVarName("len"))

		cap := preAppendContextBlock.NewGetElementPtr(input.Value, constant.NewInt(0, i32.LLVM()), constant.NewInt(1, i32.LLVM()))
		cap.SetName(getVarName("cap"))

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

		reallocatedSpaceRaw := growSliceAllocBlock.NewCall(c.externalFuncs.Realloc.LlvmFunction, castedPrevBacking, reallocSize)
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
		sliceToAppendTo = input
	} else {

		// Allocate a new slice
		newSlice := c.contextBlock.NewAlloca(input.Type.LLVM())
		len := c.contextBlock.NewGetElementPtr(newSlice, constant.NewInt(0, i32.LLVM()), constant.NewInt(0, i32.LLVM()))
		cap := c.contextBlock.NewGetElementPtr(newSlice, constant.NewInt(0, i32.LLVM()), constant.NewInt(1, i32.LLVM()))
		offset := c.contextBlock.NewGetElementPtr(newSlice, constant.NewInt(0, i32.LLVM()), constant.NewInt(2, i32.LLVM()))
		backingArray := c.contextBlock.NewGetElementPtr(newSlice, constant.NewInt(0, i32.LLVM()), constant.NewInt(3, i32.LLVM()))

		// Copy len and cap from the previous slice
		prevSliceLen := c.contextBlock.NewGetElementPtr(input.Value, constant.NewInt(0, i32.LLVM()), constant.NewInt(0, i32.LLVM()))
		prevSliceLen.SetName(getVarName("prev-len"))
		prevSliceCap := c.contextBlock.NewGetElementPtr(input.Value, constant.NewInt(0, i32.LLVM()), constant.NewInt(1, i32.LLVM()))
		prevSliceCap.SetName(getVarName("prev-cap"))

		loadedPrevLen := c.contextBlock.NewLoad(prevSliceLen)
		loadedPrevCap := c.contextBlock.NewLoad(prevSliceCap)

		c.contextBlock.NewStore(loadedPrevLen, len)
		c.contextBlock.NewStore(loadedPrevCap, cap)
		c.contextBlock.NewStore(constant.NewInt(0, i32.LLVM()), offset)

		// Allocate a new backing array, and copy the data from the previous one to the new
		// TODO: Make sure that cap is large enough for the new data
		cap64 := c.contextBlock.NewZExt(loadedPrevCap, i64.LLVM())
		mallocatedSpaceRaw := c.contextBlock.NewCall(c.externalFuncs.Malloc.LlvmFunction, cap64)
		bitcasted := c.contextBlock.NewBitCast(mallocatedSpaceRaw, llvmTypes.NewPointer(inputSlice.Type.LLVM()))
		c.contextBlock.NewStore(bitcasted, backingArray)

		// Copy data from the old backing array to the new one
		prevBackArray := c.contextBlock.NewGetElementPtr(input.Value, constant.NewInt(0, i32.LLVM()), constant.NewInt(3, i32.LLVM()))
		prevBackArray.SetName(getVarName("prev-backarr"))
		prevBackArrayLoaded := c.contextBlock.NewLoad(prevBackArray)
		prevBackArrayCasted := c.contextBlock.NewBitCast(prevBackArrayLoaded, llvmTypes.NewPointer(i8.LLVM()))
		prevBackArrayCasted.SetName(getVarName("prev-backarray-casted"))

		loadedPrevLen64 := c.contextBlock.NewZExt(loadedPrevLen, i64.LLVM())
		copyBytesSize := c.contextBlock.NewMul(loadedPrevLen64, constant.NewInt(inputSlice.Type.Size(), i64.LLVM()))

		c.contextBlock.NewCall(c.externalFuncs.Memcpy.LlvmFunction,
			// Dest
			mallocatedSpaceRaw,
			// Src
			prevBackArrayCasted,
			// n - Number of bytes to copy
			copyBytesSize,
		)

		sliceToAppendTo = value.Value{
			Type:       input.Type,
			Value:      newSlice,
			IsVariable: true,
		}

		c.contextBlock.NewBr(appendToSliceBlock)
	}

	// Add item

	// Get current len
	sliceLenPtr := appendToSliceBlock.NewGetElementPtr(sliceToAppendTo.Value, constant.NewInt(0, i32.LLVM()), constant.NewInt(0, i32.LLVM()))
	sliceLen := appendToSliceBlock.NewLoad(sliceLenPtr)
	sliceLen.SetName(getVarName("slicelen"))

	// TODO: Allow adding many items at once `foo = append(foo, bar, baz)`
	backingArrayAppendPtr := appendToSliceBlock.NewGetElementPtr(sliceToAppendTo.Value,
		constant.NewInt(0, i32.LLVM()),
		constant.NewInt(3, i32.LLVM()),
	)

	backingArrayAppendPtr.SetName(getVarName("backingarrayptr"))
	loadedPtr := appendToSliceBlock.NewLoad(backingArrayAppendPtr)
	storePtr := appendToSliceBlock.NewGetElementPtr(loadedPtr, sliceLen)

	c.contextBlock = appendToSliceBlock

	// Add type of items in slice to the context
	c.contextAssignDest = append(c.contextAssignDest, value.Value{Type: inputSlice.Type})

	addItem := c.compileValue(v.Arguments[1])

	// Convert type if neccesary
	addItem = c.valueToInterfaceValue(addItem, inputSlice.Type)

	addItemVal := addItem.Value
	if addItem.IsVariable {
		addItemVal = c.contextBlock.NewLoad(addItemVal)
	}

	// Pop assigng type stack
	c.contextAssignDest = c.contextAssignDest[0 : len(c.contextAssignDest)-1]

	appendToSliceBlock.NewStore(addItemVal, storePtr)

	// Increase len

	newLen := appendToSliceBlock.NewAdd(sliceLen, constant.NewInt(1, i32.LLVM()))
	appendToSliceBlock.NewStore(newLen, sliceLenPtr)

	c.contextBlock = appendToSliceBlock

	return sliceToAppendTo
}

func (c *Compiler) compileInitializeSliceNode(v *parser.InitializeSliceNode) value.Value {
	itemType := parserTypeToType(v.Type)

	var values []value.Value

	// Add items
	for _, val := range v.Items {
		// Push assigng type stack
		c.contextAssignDest = append(c.contextAssignDest, value.Value{Type: itemType})

		values = append(values, c.compileValue(val))

		// Pop assigng type stack
		c.contextAssignDest = c.contextAssignDest[0 : len(c.contextAssignDest)-1]
	}

	return c.compileInitializeSliceWithValues(itemType, values...)
}

func (c *Compiler) compileInitializeSliceWithValues(itemType types.Type, values ...value.Value) value.Value {
	sliceType := &types.Slice{
		Type:     itemType,
		LlvmType: internal.Slice(itemType.LLVM()),
	}

	// Create slice with cap set to the requested size
	allocSlice := sliceType.SliceZero(c.contextBlock, c.externalFuncs.Malloc.LlvmFunction, len(values))

	backingArrayPtr := c.contextBlock.NewGetElementPtr(allocSlice,
		constant.NewInt(0, i32.LLVM()),
		constant.NewInt(3, i32.LLVM()),
	)

	loadedPtr := c.contextBlock.NewLoad(backingArrayPtr)
	loadedPtr.SetName(getVarName("loadedbackingarrayptr"))

	// Add items
	for i, val := range values {
		storePtr := c.contextBlock.NewGetElementPtr(loadedPtr, constant.NewInt(int64(i), i32.LLVM()))
		storePtr.SetName(getVarName(fmt.Sprintf("storeptr-%d", i)))
		c.contextBlock.NewStore(val.Value, storePtr)
	}

	// Set len
	lenPtr := c.contextBlock.NewGetElementPtr(allocSlice,
		constant.NewInt(0, i32.LLVM()),
		constant.NewInt(0, i32.LLVM()),
	)
	c.contextBlock.NewStore(constant.NewInt(int64(len(values)), i32.LLVM()), lenPtr)

	return value.Value{
		Value:      allocSlice,
		Type:       sliceType,
		IsVariable: true,
	}
}
