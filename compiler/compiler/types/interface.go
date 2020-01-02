package types

import (
	"fmt"
	"sort"

	"github.com/llir/llvm/ir/types"
	llvmValue "github.com/llir/llvm/ir/value"
)

type Interface struct {
	backingType

	SourceName      string
	RequiredMethods map[string]InterfaceMethod
}

func (i Interface) Name() string {
	return fmt.Sprintf("interface(%s)", i.SourceName)
}

// SortedRequiredMethods returns a sorted slice of all method names
// The returned order is the order the methods will be layed out in the JumpTable
func (i Interface) SortedRequiredMethods() []string {
	var orderedMethods []string
	for methodName := range i.RequiredMethods {
		orderedMethods = append(orderedMethods, methodName)
	}
	sort.Strings(orderedMethods)
	return orderedMethods
}

func (i Interface) JumpTable() *types.StructType {
	orderedMethods := i.SortedRequiredMethods()

	var ifaceTableMethods []types.Type

	for _, methodName := range orderedMethods {
		methodSignature := i.RequiredMethods[methodName]

		var retType types.Type = types.Void
		if len(methodSignature.ReturnTypes) > 0 {
			retType = methodSignature.ReturnTypes[0].LLVM()
		}

		paramTypes := []types.Type{types.NewPointer(types.I8)}
		for _, argType := range methodSignature.ArgumentTypes {
			paramTypes = append(paramTypes, argType.LLVM())
		}

		ifaceTableMethods = append(ifaceTableMethods, types.NewPointer(types.NewFunc(retType, paramTypes...)))
	}

	return types.NewStruct(ifaceTableMethods...)
}

func (i Interface) LLVM() types.Type {
	return types.NewStruct(
		// Pointer to the backing data
		types.NewPointer(types.I8),

		// Backing data type
		types.I32,

		// Interface table
		// Used for method resolving
		types.NewPointer(i.JumpTable()),
	)
}

func (Interface) Size() int64 {
	return 64/8 * 3
}

type InterfaceMethod struct {
	backingType

	LlvmJumpFunction llvmValue.Named

	ArgumentTypes []Type
	ReturnTypes   []Type
}

func (InterfaceMethod) LLVM() types.Type {
	panic("InterfaceMethod has no LLVM value")
}

func (InterfaceMethod) Name() string {
	return "InterfaceMethod"
}
