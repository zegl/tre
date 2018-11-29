package main

import "external"

func main() {
	external.Printf("%d\n", 2 * 3 + 4) // 10
	external.Printf("%d\n", 4 + 2 * 3) // 10
}