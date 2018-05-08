package types

import (
	"fmt"
	"sort"

	"github.com/llir/llvm/ir/types"
)

type Interface struct {
	backingType

	SourceName      string
	RequiredMethods map[string]InterfaceMethod
}

func (i Interface) Name() string {
	return fmt.Sprintf("interface(%s)", i.SourceName)
}

func (i Interface) JumpTable() *types.StructType {
	var orderedMethods []string
	for methodName := range i.RequiredMethods {
		orderedMethods = append(orderedMethods, methodName)
	}
	sort.Strings(orderedMethods)

	var ifaceTableMethods []types.Type

	for _, methodName := range orderedMethods {
		methodSignature := i.RequiredMethods[methodName]

		var retType types.Type = types.Void
		if len(methodSignature.ReturnTypes) > 0 {
			retType = methodSignature.ReturnTypes[0].LLVM()
		}

		var paramTypes []*types.Param
		for _, argType := range methodSignature.ReturnTypes {
			paramTypes = append(paramTypes, types.NewParam("instance", types.NewPointer(types.I8)), types.NewParam("", argType.LLVM()))
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
	return 1 // 1 pointer
}

type InterfaceMethod struct {
	ArgumentTypes []Type
	ReturnTypes   []Type
}
