package main

import "external"

func main() {
	// 9
	external.Printf("%d\n", 77&155)

	// 223
	external.Printf("%d\n", 77|155)

	// 214
	external.Printf("%d\n", 77^155)

	// 68
	external.Printf("%d\n", 77&^155)

	// 8
	external.Printf("%d\n", 1<<3)

	// 205
	external.Printf("%d\n", 822>>2)
}
