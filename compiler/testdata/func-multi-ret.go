package main

import "external"

func otherfunc() (int, int) {
	return 100, 200
}

func main() {
	a, b := otherfunc()
	external.Printf("a: %d\n", a) // a: 100
	external.Printf("b: %d\n", b) // b: 200
}
