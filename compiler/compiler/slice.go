package compiler

import (
	"fmt"
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	llvmTypes "github.com/llir/llvm/ir/types"
	llvmValue "github.com/llir/llvm/ir/value"

	"github.com/zegl/tre/compiler/compiler/internal"
	"github.com/zegl/tre/compiler/compiler/internal/pointer"
	"github.com/zegl/tre/compiler/compiler/name"
	"github.com/zegl/tre/compiler/compiler/types"
	"github.com/zegl/tre/compiler/compiler/value"
	"github.com/zegl/tre/compiler/parser"
)

func (c *Compiler) compileSubstring(src value.Value, v *parser.SliceArrayNode) value.Value {
	srcVal := src.Value

	var originalLength *ir.InstExtractValue

	// Get backing array from string type
	if src.IsVariable {
		srcVal = c.contextBlock.NewLoad(pointer.ElemType(srcVal), srcVal)
	}
	if src.Type.Name() == "string" {
		originalLength = c.contextBlock.NewExtractValue(srcVal, 0)
		srcVal = c.contextBlock.NewExtractValue(srcVal, 1)
	}

	start := c.compileValue(v.Start)

	outsideOfLengthBr := c.contextBlock.Parent.NewBlock(name.Block())
	c.panic(outsideOfLengthBr, "Substring start larger than len")
	outsideOfLengthBr.NewUnreachable()

	safeBlock := c.contextBlock.Parent.NewBlock(name.Block())

	// Make sure that the offset is within the string length
	cmp := c.contextBlock.NewICmp(enum.IPredUGE, start.Value, originalLength)
	c.contextBlock.NewCondBr(cmp, outsideOfLengthBr, safeBlock)

	c.contextBlock = safeBlock

	offset := safeBlock.NewGetElementPtr(pointer.ElemType(srcVal), srcVal, start.Value)

	var length llvmValue.Value
	if v.HasEnd {
		length = c.compileValue(v.End).Value
	} else {
		length = constant.NewInt(llvmTypes.I64, 1)
	}

	dst := safeBlock.NewCall(c.externalFuncs.Strndup.Value.(llvmValue.Named), offset, length)

	// Convert *i8 to %string
	sType, ok := c.packages["global"].GetPkgType("string")
	if !ok {
		panic("string type not found")
	}
	alloc := c.contextBlock.NewAlloca(sType.LLVM())

	// Save length of the string
	lenItem := safeBlock.NewGetElementPtr(pointer.ElemType(alloc), alloc, constant.NewInt(llvmTypes.I32, 0), constant.NewInt(llvmTypes.I32, 0))
	lenItem.SetName(name.Var("len"))
	safeBlock.NewStore(constant.NewInt(llvmTypes.I64, 100), lenItem) // TODO

	// Save i8* version of string
	strItem := safeBlock.NewGetElementPtr(pointer.ElemType(alloc), alloc, constant.NewInt(llvmTypes.I32, 0), constant.NewInt(llvmTypes.I32, 1))
	strItem.SetName(name.Var("str"))
	safeBlock.NewStore(dst, strItem)

	return value.Value{
		Value:      safeBlock.NewLoad(pointer.ElemType(alloc), alloc),
		Type:       sType,
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
	lenItem := c.contextBlock.NewGetElementPtr(pointer.ElemType(alloc), alloc, constant.NewInt(llvmTypes.I32, 0), constant.NewInt(llvmTypes.I32, 0))
	lenItem.SetName(name.Var("len"))
	c.contextBlock.NewStore(sliceLen32, lenItem)

	// Cap
	capItem := c.contextBlock.NewGetElementPtr(pointer.ElemType(alloc), alloc, constant.NewInt(llvmTypes.I32, 0), constant.NewInt(llvmTypes.I32, 1))
	c.contextBlock.NewStore(sliceLen32, capItem)
	capItem.SetName(name.Var("cap"))

	// Offset
	offsetItem := c.contextBlock.NewGetElementPtr(pointer.ElemType(alloc), alloc, constant.NewInt(llvmTypes.I32, 0), constant.NewInt(llvmTypes.I32, 2))
	c.contextBlock.NewStore(offset32, offsetItem)
	offsetItem.SetName(name.Var("offset"))

	// Backing Array
	backingArrayItem := c.contextBlock.NewGetElementPtr(pointer.ElemType(alloc), alloc, constant.NewInt(llvmTypes.I32, 0), constant.NewInt(llvmTypes.I32, 3))
	backingArrayItem.SetName(name.Var("backing"))

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
	// 1. Grow the backing array if necessary (cap == len)
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

	// Create blocks that are needed later

	copySliceBlock := c.contextBlock.Parent.NewBlock(name.Block() + "-copy-slice")
	copySliceBlock.Term = ir.NewUnreachable()

	addToSliceBlock := c.contextBlock.Parent.NewBlock(name.Block() + "-add-to-slice")
	addToSliceBlock.Term = ir.NewUnreachable()

	appendExistingBlock := c.contextBlock.Parent.NewBlock(name.Block() + "-append-existing-block")
	appendExistingBlock.Term = ir.NewUnreachable()

	// The slice that we're appending to will be stored here
	sliceToAppendToLLVM := c.contextBlock.NewAlloca(input.Type.LLVM())
	sliceToAppendToLLVM.SetName(name.Var("sliceToAppendTo"))

	if isSelfAssign {
		preAppendContextBlock := c.contextBlock

		lenVal := preAppendContextBlock.NewGetElementPtr(pointer.ElemType(input.Value), input.Value, constant.NewInt(llvmTypes.I32, 0), constant.NewInt(llvmTypes.I32, 0))
		lenVal.SetName(name.Var("len"))

		capVal := preAppendContextBlock.NewGetElementPtr(pointer.ElemType(input.Value), input.Value, constant.NewInt(llvmTypes.I32, 0), constant.NewInt(llvmTypes.I32, 1))
		capVal.SetName(name.Var("cap"))

		loadedLen := preAppendContextBlock.NewLoad(pointer.ElemType(lenVal), lenVal)
		loadedCap := preAppendContextBlock.NewLoad(pointer.ElemType(capVal), capVal)

		shouldAppendToExisting := preAppendContextBlock.NewICmp(enum.IPredULT, loadedLen, loadedCap)

		// Add to existing backing array if len < cap
		preAppendContextBlock.NewCondBr(
			shouldAppendToExisting,
			appendExistingBlock, // append to existing backing array
			copySliceBlock,
		)
	} else {
		c.contextBlock.NewBr(copySliceBlock)
	}

	existingSliceLoaded := appendExistingBlock.NewLoad(pointer.ElemType(input.Value), input.Value)
	appendExistingBlock.NewStore(existingSliceLoaded, sliceToAppendToLLVM)
	appendExistingBlock.NewBr(addToSliceBlock)

	c.generateCopySliceBlock(copySliceBlock, addToSliceBlock, input, inputSlice, sliceToAppendToLLVM)

	c.generateAppendToSliceBlock(addToSliceBlock, sliceToAppendToLLVM, inputSlice, v)

	c.contextBlock = addToSliceBlock

	return value.Value{
		Value:      sliceToAppendToLLVM,
		Type:       inputSlice,
		IsVariable: true,
	}
}

func (c *Compiler) generateCopySliceBlock(copySliceBlock *ir.Block, appendToSliceBlock *ir.Block, input value.Value, inputSlice *types.Slice, sliceToAppendToLLVM llvmValue.Value) {
	c.contextBlock = copySliceBlock

	// Allocate a new slice
	newSlice := copySliceBlock.NewAlloca(input.Type.LLVM())
	newSlice.SetName(name.Var("copy-to-new-slice"))

	lenVal := copySliceBlock.NewGetElementPtr(pointer.ElemType(newSlice), newSlice, constant.NewInt(llvmTypes.I32, 0), constant.NewInt(llvmTypes.I32, 0))
	lenVal.SetName(name.Var("len"))

	capVal := copySliceBlock.NewGetElementPtr(pointer.ElemType(newSlice), newSlice, constant.NewInt(llvmTypes.I32, 0), constant.NewInt(llvmTypes.I32, 1))
	capVal.SetName(name.Var("cap"))

	offset := copySliceBlock.NewGetElementPtr(pointer.ElemType(newSlice), newSlice, constant.NewInt(llvmTypes.I32, 0), constant.NewInt(llvmTypes.I32, 2))
	offset.SetName(name.Var("offset"))

	backingArray := copySliceBlock.NewGetElementPtr(pointer.ElemType(newSlice), newSlice, constant.NewInt(llvmTypes.I32, 0), constant.NewInt(llvmTypes.I32, 3))
	backingArray.SetName(name.Var("backing"))

	// Copy len and cap from the previous slice
	prevSliceLen := copySliceBlock.NewGetElementPtr(pointer.ElemType(input.Value), input.Value, constant.NewInt(llvmTypes.I32, 0), constant.NewInt(llvmTypes.I32, 0))
	prevSliceLen.SetName(name.Var("prev-len"))
	prevSliceCap := copySliceBlock.NewGetElementPtr(pointer.ElemType(input.Value), input.Value, constant.NewInt(llvmTypes.I32, 0), constant.NewInt(llvmTypes.I32, 1))
	prevSliceCap.SetName(name.Var("prev-cap"))

	loadedPrevLen := copySliceBlock.NewLoad(pointer.ElemType(prevSliceLen), prevSliceLen)
	loadedPrevCap := copySliceBlock.NewLoad(pointer.ElemType(prevSliceCap), prevSliceCap)

	// Store len and offset. (The new cap has not been calculated yet)
	copySliceBlock.NewStore(loadedPrevLen, lenVal)
	copySliceBlock.NewStore(constant.NewInt(llvmTypes.I32, 0), offset)

	// Allocate a new backing array, and copy the data from the previous one to the new
	// TODO: Make sure that cap is large enough for the new data

	twiceCap := copySliceBlock.NewMul(loadedPrevCap, constant.NewInt(llvmTypes.I32, 2))
	twiceCap64 := copySliceBlock.NewZExt(twiceCap, i64.LLVM())
	sizeTimesCap := copySliceBlock.NewMul(twiceCap64, constant.NewInt(llvmTypes.I64, input.Type.Size()))
	mallocatedSpaceRaw := copySliceBlock.NewCall(c.externalFuncs.Malloc.Value.(llvmValue.Named), sizeTimesCap)
	mallocatedSpaceRaw.SetName(name.Var("slice-grow"))

	// Store new cap
	copySliceBlock.NewStore(twiceCap, capVal)

	bitcasted := copySliceBlock.NewBitCast(mallocatedSpaceRaw, llvmTypes.NewPointer(inputSlice.Type.LLVM()))
	copySliceBlock.NewStore(bitcasted, backingArray)

	// Copy data from the old backing array to the new one
	prevBackArray := copySliceBlock.NewGetElementPtr(pointer.ElemType(input.Value), input.Value, constant.NewInt(llvmTypes.I32, 0), constant.NewInt(llvmTypes.I32, 3))
	prevBackArray.SetName(name.Var("prev-backarr"))

	prevBackArrayLoaded := copySliceBlock.NewLoad(pointer.ElemType(prevBackArray), prevBackArray)
	prevBackArrayCasted := copySliceBlock.NewBitCast(prevBackArrayLoaded, llvmTypes.NewPointer(i8.LLVM()))
	prevBackArrayCasted.SetName(name.Var("prev-backarray-casted"))

	copyIndex := copySliceBlock.NewAlloca(llvmTypes.I32)
	copySliceBlock.NewStore(constant.NewInt(llvmTypes.I32, 0), copyIndex)

	loadedNewSlice := copySliceBlock.NewLoad(pointer.ElemType(newSlice), newSlice)
	copySliceBlock.NewStore(loadedNewSlice, sliceToAppendToLLVM)

	// Copy all items, one by one

	copyBlock := copySliceBlock.Parent.NewBlock(name.Block() + "-copy-slice-bytes")
	prevArrItemPtr := copyBlock.NewGetElementPtr(pointer.ElemType(prevBackArrayLoaded), prevBackArrayLoaded, copyBlock.NewLoad(pointer.ElemType(copyIndex), copyIndex))
	newArrItemPtr := copyBlock.NewGetElementPtr(pointer.ElemType(bitcasted), bitcasted, copyBlock.NewLoad(pointer.ElemType(copyIndex), copyIndex))
	copyBlock.NewStore(copyBlock.NewLoad(pointer.ElemType(prevArrItemPtr), prevArrItemPtr), newArrItemPtr)
	a := copyBlock.NewAdd(constant.NewInt(llvmTypes.I32, 1), copyBlock.NewLoad(pointer.ElemType(copyIndex), copyIndex))
	copyBlock.NewStore(a, copyIndex)
	cmp := copyBlock.NewICmp(enum.IPredULT, a, loadedPrevLen)
	copyBlock.NewCondBr(cmp, copyBlock, appendToSliceBlock)

	copySliceBlock.NewBr(copyBlock)
}

func (c *Compiler) generateAppendToSliceBlock(appendToSliceBlock *ir.Block, sliceToAppendTo llvmValue.Value, inputSlice *types.Slice, v *parser.CallNode) {
	c.contextBlock = appendToSliceBlock

	// Add item

	// Get current len
	sliceLenPtr := appendToSliceBlock.NewGetElementPtr(pointer.ElemType(sliceToAppendTo), sliceToAppendTo, constant.NewInt(llvmTypes.I32, 0), constant.NewInt(llvmTypes.I32, 0))
	sliceLenPtr.SetName(name.Var("sliceLenPtr"))

	sliceLen := appendToSliceBlock.NewLoad(pointer.ElemType(sliceLenPtr), sliceLenPtr)
	sliceLen.SetName(name.Var("slicelen"))

	// TODO: Allow adding many items at once `foo = append(foo, bar, baz)`
	backingArrayAppendPtr := appendToSliceBlock.NewGetElementPtr(pointer.ElemType(sliceToAppendTo), sliceToAppendTo,
		constant.NewInt(llvmTypes.I32, 0),
		constant.NewInt(llvmTypes.I32, 3),
	)

	backingArrayAppendPtr.SetName(name.Var("backingarrayptr"))
	loadedPtr := appendToSliceBlock.NewLoad(pointer.ElemType(backingArrayAppendPtr), backingArrayAppendPtr)
	storePtr := appendToSliceBlock.NewGetElementPtr(pointer.ElemType(loadedPtr), loadedPtr, sliceLen)
	storePtr.SetName(name.Var("store-ptr"))

	// Add type of items in slice to the context
	c.contextAssignDest = append(c.contextAssignDest, value.Value{Type: inputSlice.Type})

	addItem := c.compileValue(v.Arguments[1])

	// Convert type if necessary
	addItem = c.valueToInterfaceValue(addItem, inputSlice.Type)

	addItemVal := addItem.Value
	if addItem.IsVariable {
		addItemVal = appendToSliceBlock.NewLoad(pointer.ElemType(addItemVal), addItemVal)
	}

	// Pop assigning type stack
	c.contextAssignDest = c.contextAssignDest[0 : len(c.contextAssignDest)-1]

	appendToSliceBlock.NewStore(addItemVal, storePtr)

	// Increase len

	newLen := appendToSliceBlock.NewAdd(sliceLen, constant.NewInt(llvmTypes.I32, 1))
	appendToSliceBlock.NewStore(newLen, sliceLenPtr)
}

func (c *Compiler) compileInitializeSliceNode(v *parser.InitializeSliceNode) value.Value {
	itemType := c.parserTypeToType(v.Type)

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
	allocSlice := sliceType.SliceZero(c.contextBlock, c.externalFuncs.Malloc.Value.(llvmValue.Named), len(values))

	backingArrayPtr := c.contextBlock.NewGetElementPtr(pointer.ElemType(allocSlice), allocSlice,
		constant.NewInt(llvmTypes.I32, 0),
		constant.NewInt(llvmTypes.I32, 3),
	)

	loadedPtr := c.contextBlock.NewLoad(pointer.ElemType(backingArrayPtr), backingArrayPtr)
	loadedPtr.SetName(name.Var("loadedbackingarrayptr"))

	// Add items
	for i, val := range values {
		storePtr := c.contextBlock.NewGetElementPtr(pointer.ElemType(loadedPtr), loadedPtr, constant.NewInt(llvmTypes.I32, int64(i)))
		storePtr.SetName(name.Var(fmt.Sprintf("storeptr-%d", i)))

		val = c.valueToInterfaceValue(val, itemType)
		v := val.Value
		if val.IsVariable {
			v = c.contextBlock.NewLoad(pointer.ElemType(v), v)
		}
		c.contextBlock.NewStore(v, storePtr)
	}

	// Set len
	lenPtr := c.contextBlock.NewGetElementPtr(pointer.ElemType(allocSlice), allocSlice,
		constant.NewInt(llvmTypes.I32, 0),
		constant.NewInt(llvmTypes.I32, 0),
	)
	lenPtr.SetName(name.Var("len"))
	c.contextBlock.NewStore(constant.NewInt(llvmTypes.I32, int64(len(values))), lenPtr)

	return value.Value{
		Value:      allocSlice,
		Type:       sliceType,
		IsVariable: true,
	}
}
