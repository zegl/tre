package main

import "external"

const offset = 5

var global = 1000

func main() {
	x := 10
	s := "hello"
	f := func() int {
		return x + offset + global
	}
	g := func() {
		external.Printf("%s %d\n", s, x)
	}

	external.Printf("%d\n", f()) // 1015
	g()                          // hello 10

	x = 20
	external.Printf("%d\n", f()) // 1025
}
