package parser

import (
	"fmt"
	"log"

	"github.com/zegl/tre/compiler/lexer"
)

type SwitchNode struct {
	baseNode
	Item        Node
	Cases       []SwitchCaseNode
	DefaultBody Node // can be null
}

type SwitchCaseNode struct {
	Condition Node
	Body      []Node
}

func (s SwitchNode) String() string {
	return fmt.Sprintf("switch %s", s.Item)
}

func (p *parser) parseSwitch() *SwitchNode {
	p.i++

	s := &SwitchNode{
		Item:  p.parseOne(true),
		Cases: make([]SwitchCaseNode, 0),
	}

	p.i++
	p.expect(p.lookAhead(0), lexer.Item{Type: lexer.SEPARATOR, Val: "{"})
	p.i++

	for {
		next := p.lookAhead(0)
		log.Printf("next: %+v", next)

		if next.Type == lexer.EOL {
			p.i++
			continue
		}

		if next.Type == lexer.KEYWORD && next.Val == "case" {
			p.i++
			switchCase := SwitchCaseNode{
				Condition: p.parseOne(true),
			}

			p.i++
			p.expect(p.lookAhead(0), lexer.Item{Type: lexer.OPERATOR, Val: ":"})
			p.i++

			var reached lexer.Item
			switchCase.Body, reached = p.parseUntilEither(
				[]lexer.Item{
					{Type: lexer.SEPARATOR, Val: "}"},
					{Type: lexer.KEYWORD, Val: "case"},
					{Type: lexer.KEYWORD, Val: "default"},
					{Type: lexer.KEYWORD, Val: "fallthrough"},
				},
			)

			s.Cases = append(s.Cases, switchCase)

			// reached end of switch
			if reached.Type == lexer.SEPARATOR && reached.Val == "}" {
				break
			}
		}

		// TODO: default
		// TODO: fallthrough
	}

	return s
}
