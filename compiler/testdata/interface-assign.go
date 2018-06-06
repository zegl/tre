package main

import "external"

type myInterfaceType struct {
	A interface{}
}

func main() {
	var target interface{}

	target = 123
	realInt, ok := target.(int64)
	external.Printf("%d %d\n", ok, realInt) // 1 123

	target = "hello"
	realString, ok := target.(string)
	external.Printf("%d %s\n", ok, realString) // 1 hello

	foo := 456
	target = foo
	realVarInt, ok := target.(int64)
	external.Printf("%d %d\n", ok, realVarInt) // 1 456

	var targetSlice []interface{}
	targetSlice = append(targetSlice, 789)
	realSliceInt, ok := targetSlice[0].(int64)
	external.Printf("%d %d\n", ok, realSliceInt) // 1 789

	var targetStruct myInterfaceType
	targetStruct.A = 654
	realStructTarget, ok := targetStruct.A.(int64)
	external.Printf("%d %d\n", ok, realStructTarget) // 1 654
}
