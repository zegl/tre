package main

import "external"

func main() {
	i := 0
	i++
	external.Printf("%d\n", i) // 1
	external.Printf("%d\n", i++) // 2
	external.Printf("%d\n", i--) // 1
	i--
	external.Printf("%d\n", i) // 0
}
