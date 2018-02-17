package types

import (
	"fmt"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/types"
)

type Type interface {
	LLVM() types.Type
	Name() string

	AddMethod(string, *Method)
	GetMethod(string) (*Method, bool)
}

type backingType struct {
	methods map[string]*Method
}

func (b *backingType) AddMethod(name string, method *Method) {
	if b.methods == nil {
		b.methods = make(map[string]*Method)
	}
	b.methods[name] = method
}

func (b *backingType) GetMethod(name string) (*Method, bool) {
	m, ok := b.methods[name]
	return m, ok
}

type Struct struct {
	backingType

	Members       map[string]Type
	MemberIndexes map[string]int

	SourceName string
	Type       types.Type
}

func (s Struct) LLVM() types.Type {
	return s.Type
}

func (s Struct) Name() string {
	return fmt.Sprintf("struct(%s)", s.SourceName)
}

type Method struct {
	backingType

	Function        *Function
	PointerReceiver bool
	MethodName      string
}

func (m Method) LLVM() types.Type {
	return m.Function.LLVM()
}

func (m Method) Name() string {
	return m.MethodName
}

type Function struct {
	backingType

	LlvmFunction *ir.Function
	ReturnType   Type
	FunctionName string
}

func (f Function) LLVM() types.Type {
	return f.LlvmFunction.Type()
}

func (f Function) Name() string {
	return f.FunctionName
}

type Basic struct {
	backingType

	Type     types.Type
	TypeName string
}

func (b Basic) LLVM() types.Type {
	return b.Type
}

func (b Basic) Name() string {
	return b.TypeName
}

type StringType struct {
	backingType
	Type types.Type
}

// Populated by compiler.go
var ModuleStringType types.Type

func (s StringType) LLVM() types.Type {
	return ModuleStringType
}

func (s StringType) Name() string {
	return "string"
}

type Array struct {
	backingType
	Type     Type
	LlvmType types.Type
}

func (a Array) LLVM() types.Type {
	return a.LlvmType
}

func (a Array) Name() string {
	return "array"
}
