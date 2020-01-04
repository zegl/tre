package compiler

import (
	"github.com/zegl/tre/compiler/compiler/value"
)

// Representation of a Go package
type pkg struct {
	name string
	vars map[string]value.Value
}

func NewPkg(name string) *pkg {
	return &pkg{
		name: name,
	}
}

func (p *pkg) DefinePkgVar(name string, val value.Value) {
	if p.vars == nil {
		p.vars = make(map[string]value.Value)
	}
	p.vars[name] = val

}

func (p *pkg) GetPkgVar(name string) (value.Value, bool) {
	v, ok := p.vars[name]
	return v, ok
}
