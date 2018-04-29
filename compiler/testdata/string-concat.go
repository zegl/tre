package main

import "external"

func main() {
	// 3 - foo
	foo := "foo"
	external.Printf("%d - %s\n", len(foo), foo)

	// 10 - fooyollllo
	foo = foo + "yollllo"
	external.Printf("%d - %s\n", len(foo), foo)

	// 16 - fooyolllloh11100
	foo = foo + "h11100"
	external.Printf("%d - %s\n", len(foo), foo)

	// abbcccdddd
	external.Printf("%s\n", "a"+"bb"+"ccc"+"dddd")
}
