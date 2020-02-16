package const_iota

import (
	"log"

	"github.com/zegl/tre/compiler/parser"
)

func Iota(root *parser.FileNode) *parser.FileNode {
	parser.Walk(&iotaVisitor{}, root)
	return root
}

type iotaVisitor struct {
	count int64
}

func (i *iotaVisitor) Visit(node parser.Node) (w parser.Visitor) {
	w = i

	if _, ok := node.(*parser.AllocGroup); ok {
		i.count = 0
	}

	if a, ok := node.(*parser.AllocNode); ok {
		if !a.IsConst {
			return
		}

		if len(a.Val) != 1 {
			return
		}

		if n, ok := a.Val[0].(*parser.NameNode); ok {
			if n.Name == "iota" {
				log.Println("found iota!", i.count)

				a.Val[0] = &parser.ConstantNode{
					Type:  parser.NUMBER,
					Value: i.count,
				}

				i.count++
			}
		}
	}

	return
}
