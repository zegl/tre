package main

import "external"

func main() {
	i := 100
	iptr1 := &i
	iptr2 := &iptr1
	iptr3 := &iptr2

	// i: 100
	// *iptr1: 100
	// **iptr2: 100
	// ***iptr3: 100
	external.Printf("i: %d\n", i)
	external.Printf("*iptr1: %d\n", *iptr1)
	external.Printf("**iptr2: %d\n", **iptr2)
	external.Printf("***iptr3: %d\n", ***iptr3)
}
