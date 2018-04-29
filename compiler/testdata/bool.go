package main

import "external"

func main() {
	// 1
	// 0
	// a was true
	// b was not true
	// 0

	a := true
	external.Printf("%d\n", a)

	a = false
	external.Printf("%d\n", a)

	a = true
	if a {
		println("a was true")
	}
	if !a {
		println("a was not true")
	}

	b := false
	if b {
		println("b was true")
	}
	if !b {
		println("b was not true")
	}

	var c bool
	c = false
	external.Printf("%d\n", c)
}
