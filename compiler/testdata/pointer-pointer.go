package main

import "external"

func main() {
	i := 100
	iptr1 := &i
	iptr2 := &iptr1
	iptr3 := &iptr2

	external.Printf("i: %d\n", i)               // i: 100
	external.Printf("*iptr1: %d\n", *iptr1)     // *iptr1: 100
	external.Printf("**iptr2: %d\n", **iptr2)   // **iptr2: 100
	external.Printf("***iptr3: %d\n", ***iptr3) // ***iptr3: 100

	i = 200
	external.Printf("i: %d\n", i)               // i: 200
	external.Printf("*iptr1: %d\n", *iptr1)     // *iptr1: 200
	external.Printf("**iptr2: %d\n", **iptr2)   // **iptr2: 200
	external.Printf("***iptr3: %d\n", ***iptr3) // ***iptr3: 200

	*iptr1 = 300
	external.Printf("i: %d\n", i)               // i: 300
	external.Printf("*iptr1: %d\n", *iptr1)     // *iptr1: 300
	external.Printf("**iptr2: %d\n", **iptr2)   // **iptr2: 300
	external.Printf("***iptr3: %d\n", ***iptr3) // ***iptr3: 300

	**iptr2 = 400
	external.Printf("i: %d\n", i)               // i: 400
	external.Printf("*iptr1: %d\n", *iptr1)     // *iptr1: 400
	external.Printf("**iptr2: %d\n", **iptr2)   // **iptr2: 400
	external.Printf("***iptr3: %d\n", ***iptr3) // ***iptr3: 400

	***iptr3 = 500
	external.Printf("i: %d\n", i)               // i: 500
	external.Printf("*iptr1: %d\n", *iptr1)     // *iptr1: 500
	external.Printf("**iptr2: %d\n", **iptr2)   // **iptr2: 500
	external.Printf("***iptr3: %d\n", ***iptr3) // ***iptr3: 500

}
