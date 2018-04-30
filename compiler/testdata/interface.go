package main

import "external"

func foo(bar interface{}) {
	real, ok := bar.(int64)
	if ok {
		external.Printf("%d\n", real)
	}

	external.Printf("after\n")
}

func main() {
	// 123
	// after
	foo(123)

	// after
	foo(false)
}
