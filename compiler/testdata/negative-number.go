package main

import "external"


func baz() int32 {
	return int32(-18)
}


func main() {
	foo := -4
	// foo is = -4
	// -foo is = 4
	external.Printf("foo is = %d\n", foo)
	external.Printf("-foo is = %d\n", -foo)

	// bar is = -16
	a := 16
	bar := -a
	external.Printf("bar is = %d\n", bar)

	// baz is = -18
	external.Printf("baz is = %d\n", baz())

	// -baz is = 18
	external.Printf("-baz is = %d\n", -baz())
}
