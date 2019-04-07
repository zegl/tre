package main

import "external"

func swap(x string, y string) (string, string) {
	return y, x
}

func main() {
	// world hello
	a, b := swap("hello", "world")
	external.Printf("%s %s\n", a, b)
}
