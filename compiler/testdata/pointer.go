package main

import "external"

func main() {
	i := 100
	iptr := &i

	// 100
	// 100
	external.Printf("%d\n", i)
	external.Printf("%d\n", *iptr)

	*iptr = 200

	// 200
	// 200
	external.Printf("%d\n", i)
	external.Printf("%d\n", *iptr)

	i = 300

	// 300
	// 300
	external.Printf("%d\n", i)
	external.Printf("%d\n", *iptr)
}
