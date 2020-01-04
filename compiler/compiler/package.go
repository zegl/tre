package compiler

import (
	"github.com/zegl/tre/compiler/compiler/types"
	"github.com/zegl/tre/compiler/compiler/value"
)

// Representation of a Go package
type pkg struct {
	name  string
	vars  map[string]value.Value
	types map[string]types.Type
}

func NewPkg(name string) *pkg {
	return &pkg{
		name:  name,
		vars:  make(map[string]value.Value),
		types: make(map[string]types.Type),
	}
}

func (p *pkg) DefinePkgVar(name string, val value.Value) {
	p.vars[name] = val
}

func (p *pkg) GetPkgVar(name string) (value.Value, bool) {
	v, ok := p.vars[name]
	return v, ok
}

func (p *pkg) DefinePkgType(name string, ty types.Type) {
	p.types[name] = ty
}

func (p *pkg) GetPkgType(name string) (types.Type, bool) {
	v, ok := p.types[name]
	return v, ok
}
