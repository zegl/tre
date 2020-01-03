package parser

import (
	"fmt"
	"github.com/zegl/tre/compiler/lexer"
)

func (p *parser) parseImport() *ImportNode {
	p.i++


	// Single import statement
	expectPathString := p.lookAhead(0)
	if expectPathString.Type == lexer.STRING {
		return &ImportNode{
			PackagePaths: []string{expectPathString.Val},
		}
	}

	// Multiple imports
	p.expect(lexer.Item{Type: lexer.OPERATOR, Val: "("}, p.lookAhead(0))
	p.i++

	var imports []string

	for {
		checkIfEndParen := p.lookAhead(0)
		if checkIfEndParen.Type == lexer.OPERATOR && checkIfEndParen.Val == ")" {
			break
		}
		if checkIfEndParen.Type == lexer.EOL {
			p.i++
			continue
		}

		if checkIfEndParen.Type == lexer.STRING {
			imports = append(imports, checkIfEndParen.Val)
			p.i++
			continue
		}

		panic(fmt.Sprintf("Failed to parse import: %+v", checkIfEndParen))
	}

	return &ImportNode{
		PackagePaths: imports,
	}
}
