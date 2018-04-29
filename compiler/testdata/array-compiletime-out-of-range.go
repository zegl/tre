package main  import "external"

import "external"

func main() {
	var array [4]int

	// compile panic: index out of range
	external.external.Printf("%d\n", array[10])
}
