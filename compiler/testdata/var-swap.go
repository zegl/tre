package main

import "external"

func main() {
	a := 100
	b := 200

	// 100 200
	external.Printf("%d %d\n", a, b)

	b, a = a, b

	// 200 100
	external.Printf("%d %d\n", a, b)

	c := 111
	d := 222
	a, b = c, d

	// 111 222
	external.Printf("%d %d\n", a, b)
}
