package compiler

import (
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	llvmTypes "github.com/llir/llvm/ir/types"
	llvmValue "github.com/llir/llvm/ir/value"
	"github.com/zegl/tre/compiler/compiler/types"
	"github.com/zegl/tre/compiler/compiler/value"
	"github.com/zegl/tre/compiler/parser"
)

func (c *Compiler) compileLoadArrayElement(v *parser.LoadArrayElement) value.Value {
	arr := c.compileValue(v.Array)
	arrayValue := arr.Value

	index := c.compileValue(v.Pos)
	indexVal := index.Value
	if index.IsVariable {
		indexVal = c.contextBlock.NewLoad(indexVal)
	}

	var runtimeLength llvmValue.Value
	var compileTimeLength uint64
	lengthKnownAtCompileTime := false
	lengthKnownAtRunTime := false
	isLlvmArrayBased := false
	var retType types.Type

	// Length of array
	if arr, ok := arr.Type.(*types.Array); ok {
		if arrayType, ok := arr.LlvmType.(*llvmTypes.ArrayType); ok {
			compileTimeLength = arrayType.Len
			lengthKnownAtCompileTime = true
			retType = arr.Type
			isLlvmArrayBased = true
		}
	}

	// Length of slice
	if slice, ok := arr.Type.(*types.Slice); ok {
		lengthKnownAtCompileTime = false
		lengthKnownAtRunTime = true

		// When the slice is a slice window into an array
		var isSliceOfArray bool

		retType = slice.Type

		arrayValue = c.contextBlock.NewLoad(arrayValue)

		// One more load is needed when the slice is a window into a LLVM array
		if _, ok := arrayValue.Type().(*llvmTypes.PointerType); ok {
			isSliceOfArray = true
		}

		if isSliceOfArray {
			arrayValue = c.contextBlock.NewLoad(arrayValue)
		}

		indexVal = c.contextBlock.NewTrunc(indexVal, i32.LLVM())

		sliceValue := arrayValue

		// Length of the slice
		runtimeLength = c.contextBlock.NewExtractValue(sliceValue, 0)

		// Add offset to indexVal
		backingArrayOffset := c.contextBlock.NewExtractValue(sliceValue, 2)
		indexVal = c.contextBlock.NewAdd(indexVal, backingArrayOffset)

		// Add offset to runtimeLength
		runtimeLength = c.contextBlock.NewAdd(runtimeLength, backingArrayOffset)

		// Backing array
		arrayValue = c.contextBlock.NewExtractValue(sliceValue, 3)
	}

	// Length of string
	if !lengthKnownAtCompileTime {
		// Get backing array from string type
		if arr.Type.Name() == "string" {
			if arr.IsVariable {
				arrayValue = c.contextBlock.NewLoad(arrayValue)
			}

			runtimeLength = c.contextBlock.NewExtractValue(arrayValue, 0)
			// Get backing array
			arrayValue = c.contextBlock.NewExtractValue(arrayValue, 1)
			lengthKnownAtRunTime = true
			retType = types.I8
			isLlvmArrayBased = false
		}
	}

	if !lengthKnownAtCompileTime && !lengthKnownAtRunTime {
		panic("unable to LoadArrayElement: could not calculate max length")
	}

	isCheckedAtCompileTime := false

	if lengthKnownAtCompileTime {
		if compileTimeLength < 0 {
			compilePanic("index out of range")
		}

		if intType, ok := index.Value.(*constant.Int); ok {
			if intType.X.IsInt64() {
				isCheckedAtCompileTime = true

				if intType.X.Uint64() > compileTimeLength {
					compilePanic("index out of range")
				}
			}
		}
	}

	if !isCheckedAtCompileTime {
		outsideOfLengthBlock := c.contextBlock.Parent.NewBlock(getBlockName() + "-array-index-out-of-range")
		c.panic(outsideOfLengthBlock, "index out of range")
		outsideOfLengthBlock.NewUnreachable()

		safeBlock := c.contextBlock.Parent.NewBlock(getBlockName() + "-after-array-index-check")

		var runtimeOrCompiletimeCmp *ir.InstICmp
		if lengthKnownAtCompileTime {
			runtimeOrCompiletimeCmp = c.contextBlock.NewICmp(enum.IPredSGE, indexVal, constant.NewInt(llvmTypes.I32, int64(compileTimeLength)))
		} else {
			runtimeOrCompiletimeCmp = c.contextBlock.NewICmp(enum.IPredSGE, indexVal, runtimeLength)
		}

		outOfRangeCmp := c.contextBlock.NewOr(
			c.contextBlock.NewICmp(enum.IPredSLT, indexVal, constant.NewInt(llvmTypes.I64, 0)),
			runtimeOrCompiletimeCmp,
		)

		c.contextBlock.NewCondBr(outOfRangeCmp, outsideOfLengthBlock, safeBlock)

		c.contextBlock = safeBlock
	}

	var indicies []llvmValue.Value
	if isLlvmArrayBased {
		indicies = append(indicies, constant.NewInt(llvmTypes.I64, 0))
	}
	indicies = append(indicies, indexVal)

	return value.Value{
		Value:      c.contextBlock.NewGetElementPtr(arrayValue, indicies...),
		Type:       retType,
		IsVariable: true,
	}
}
