package main

import "external"

func main() {
	foo := 4
	foo = foo + 5
	foo = foo + 6 + 7 + 8
	// foo is = 30
	external.Printf("foo is = %d\n", foo)
}
