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
	DefaultBody []Node // can be null
}

type SwitchCaseNode struct {
	Conditions  []Node
	Body        []Node
	Fallthrough bool
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
				Conditions: []Node{p.parseOne(true)},
			}

			p.i++

			for {
				curr := p.lookAhead(0)
				if curr.Type == lexer.OPERATOR && curr.Val == ":" {
					p.i++
					break
				}

				if curr.Type == lexer.SEPARATOR && curr.Val == "," {
					p.i++
					switchCase.Conditions = append(append(switchCase.Conditions,
						p.parseOne(true),
					))
					p.i++
					continue
				}

				panic(fmt.Sprintf("Expected : or , in case. Got %+v", curr))
			}

			var reached lexer.Item
			switchCase.Body, reached = p.parseUntilEither(
				[]lexer.Item{
					{Type: lexer.SEPARATOR, Val: "}"},
					{Type: lexer.KEYWORD, Val: "case"},
					{Type: lexer.KEYWORD, Val: "default"},
					{Type: lexer.KEYWORD, Val: "fallthrough"},
				},
			)

			if reached.Type == lexer.KEYWORD && reached.Val == "fallthrough" {
				switchCase.Fallthrough = true
				p.i++
			}

			s.Cases = append(s.Cases, switchCase)

			// reached end of switch
			if reached.Type == lexer.SEPARATOR && reached.Val == "}" {
				break
			}
		}

		if next.Type == lexer.KEYWORD && next.Val == "default" {
			p.i++
			p.expect(p.lookAhead(0), lexer.Item{Type: lexer.OPERATOR, Val: ":"})
			p.i++

			body, reached := p.parseUntilEither(
				[]lexer.Item{
					{Type: lexer.SEPARATOR, Val: "}"},
					{Type: lexer.KEYWORD, Val: "case"},
				},
			)

			s.DefaultBody = body

			// reached end of switch
			if reached.Type == lexer.SEPARATOR && reached.Val == "}" {
				break
			}
		}
	}

	log.Printf("%+v", s.Cases)

	return s
}
