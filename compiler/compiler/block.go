package compiler

import "fmt"

var blockIndex uint64

func getBlockName() string {
	name := fmt.Sprintf("block-%d", blockIndex)
	blockIndex++
	return name
}
