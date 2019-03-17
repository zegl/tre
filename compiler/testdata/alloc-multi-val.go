package main

import "external"

func main() {
	// 10 20
	a1, b1 := 10, 20
	external.Printf("%d %d\n", a1, b1)

	// 10 20 30
	a2, b2, c2 := 10, 20, 30
	external.Printf("%d %d %d\n", a2, b2, c2)
}
