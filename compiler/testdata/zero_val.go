package main

import "external"

func main() {
	var b bool
	external.Printf("%d\n", b) // 0

	var i8 int8
	external.Printf("%d\n", i8) // 0

	var i32 int
	external.Printf("%d\n", i32) // 0
}
