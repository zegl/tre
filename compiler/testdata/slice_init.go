package main

import "external"

func main() {
	foo := []int{1, 2, 3}
	external.Printf("%d\n", foo[0]) // 1
	external.Printf("%d\n", foo[1]) // 2
	external.Printf("%d\n", foo[2]) // 3

	external.Printf("%d\n", len(foo)) // 3
	external.Printf("%d\n", cap(foo)) // 3

	bar := []int{}
	external.Printf("%d\n", len(bar)) // 0
	external.Printf("%d\n", cap(bar)) // 2
}
