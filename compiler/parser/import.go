package parser

import (
	"github.com/zegl/tre/compiler/lexer"
)

func (p *parser) parseImport() ImportNode {
	p.i++

	expectPathString := p.lookAhead(0)
	if expectPathString.Type != lexer.STRING {
		panic("Expected string after import")
	}

	return ImportNode{
		PackagePath: expectPathString.Val,
	}
}
