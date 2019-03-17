package compiler

import "fmt"

var blockIndex uint64
var anonFuncIndex uint64

func getBlockName() string {
	name := fmt.Sprintf("block-%d", blockIndex)
	blockIndex++
	return name
}

func getAnonFuncName() string {
	name := fmt.Sprintf("fn%d", anonFuncIndex)
	anonFuncIndex++
	return name
}

func getVarName(prefix string) string {
	name := fmt.Sprintf("%s%d", prefix, blockIndex)
	blockIndex++
	return name
}
