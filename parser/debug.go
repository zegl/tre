package parser

import "fmt"

func (p *parser) printInput() {
	for i, p := range p.input {
		fmt.Printf("%d - %+v\n", i, p)
	}
}
