package main

import "external"

func main() {
	// 0 0 0 0
	a := [4]int{}
	external.Printf("%d %d %d %d\n", a[0], a[1], a[2], a[3])

	// 10 20 30 40
	b := [4]int{10, 20, 30, 40}
	external.Printf("%d %d %d %d\n", b[0], b[1], b[2], b[3])

	// 100 0 0 0
	c := [4]int{100}
	external.Printf("%d %d %d %d\n", c[0], c[1], c[2], c[3])
}
