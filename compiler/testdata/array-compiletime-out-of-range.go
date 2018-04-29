package main

import "external"

func main() {
	var array [4]int

	// compile panic: index out of range
	external.Printf("%d\n", array[10])
}
