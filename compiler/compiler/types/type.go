package types

import (
	"fmt"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
)

type Type interface {
	LLVM() types.Type
	Name() string

	// Size of type in bytes
	Size() int64

	AddMethod(string, *Method)
	GetMethod(string) (*Method, bool)

	Zero(*ir.BasicBlock, *ir.InstAlloca)
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

func (backingType) Size() int64 {
	panic("Type does not have size set")
}

func (backingType) Zero(*ir.BasicBlock, *ir.InstAlloca) {
	// NOOP
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
	IsExternal   bool
}

func (f Function) LLVM() types.Type {
	return f.LlvmFunction.Type()
}

func (f Function) Name() string {
	return f.FunctionName
}

type BoolType struct {
	backingType
}

func (BoolType) LLVM() types.Type {
	return types.I1
}

func (BoolType) Name() string {
	return "bool"
}

func (BoolType) Size() int64 {
	return 1
}

func (b BoolType) Zero(block *ir.BasicBlock, alloca *ir.InstAlloca) {
	block.NewStore(constant.NewInt(0, b.LLVM()), alloca)
}

type VoidType struct {
	backingType
}

func (VoidType) LLVM() types.Type {
	return types.Void
}

func (VoidType) Name() string {
	return "void"
}

func (VoidType) Size() int64 {
	return 0
}

type Int struct {
	backingType

	Type     types.Type
	TypeName string
	TypeSize int64
}

func (i Int) LLVM() types.Type {
	return i.Type
}

func (i Int) Name() string {
	return i.TypeName
}

func (i Int) Size() int64 {
	return i.TypeSize
}

func (i Int) Zero(block *ir.BasicBlock, alloca *ir.InstAlloca) {
	block.NewStore(constant.NewInt(0, i.Type), alloca)
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

type Slice struct {
	backingType
	Type     Type
	LlvmType types.Type
}

func (s Slice) LLVM() types.Type {
	return s.LlvmType
}

func (Slice) Name() string {
	return "slice"
}

func (Slice) Size() int64 {
	return 3*4 + 4 // 3 int32s and a pointer
}

func (s Slice) SliceZero(block *ir.BasicBlock, mallocFunc *ir.Function, initCap int) *ir.InstAlloca {
	// The cap must always be larger than 0
	// Use 2 as the default value
	if initCap < 2 {
		initCap = 2
	}

	emptySlize := block.NewAlloca(s.LLVM())

	len := block.NewGetElementPtr(emptySlize, constant.NewInt(0, types.I32), constant.NewInt(0, types.I32))
	cap := block.NewGetElementPtr(emptySlize, constant.NewInt(0, types.I32), constant.NewInt(1, types.I32))
	offset := block.NewGetElementPtr(emptySlize, constant.NewInt(0, types.I32), constant.NewInt(2, types.I32))
	backingArray := block.NewGetElementPtr(emptySlize, constant.NewInt(0, types.I32), constant.NewInt(3, types.I32))

	block.NewStore(constant.NewInt(0, types.I32), len)
	block.NewStore(constant.NewInt(int64(initCap), types.I32), cap)
	block.NewStore(constant.NewInt(0, types.I32), offset)

	mallocatedSpaceRaw := block.NewCall(mallocFunc, constant.NewInt(int64(initCap)*s.Type.Size(), types.I64))
	bitcasted := block.NewBitCast(mallocatedSpaceRaw, types.NewPointer(s.Type.LLVM()))
	block.NewStore(bitcasted, backingArray)

	return emptySlize
}

type Pointer struct {
	backingType

	Type     Type
	LlvmType types.Type
}

func (p Pointer) LLVM() types.Type {
	return p.LlvmType
}

func (p Pointer) Name() string {
	return "pointer"
}
