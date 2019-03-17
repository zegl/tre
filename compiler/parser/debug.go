package parser

import (
	"log"
)

func (p *parser) printInput() {
	for i, p := range p.input {
		log.Printf("%d - %+v\n", i, p)
	}
}
