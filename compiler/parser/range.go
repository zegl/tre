package parser

import (
	"fmt"
)

type RangeNode struct {
	baseNode
	Item Node
}

func (r RangeNode) String() string {
	return fmt.Sprintf("range %s", r.Item)
}

func (p *parser) parseRange() *RangeNode {
	p.i++
	return &RangeNode{Item: p.parseOne(true)}
}
