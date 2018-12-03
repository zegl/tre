package main

import "external"

func f1() int {
	return 100
}

func f2(a int, b int) int {
	return a + b
}

func main() {
	fn := f1
	external.Printf("%d\n", fn()) // 100

	ff := f2
	external.Printf("%d\n", ff(5, 6)) // 11
}
