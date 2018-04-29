package main

import "external"
import "fmt"

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
		fmt.Println("a was true")
	}
	if !a {
		fmt.Println("a was not true")
	}

	b := false
	if b {
		fmt.Println("b was true")
	}
	if !b {
		fmt.Println("b was not true")
	}

	var c bool
	c = false
	external.Printf("%d\n", c)
}
