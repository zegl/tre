package main

import "external"

func main() {
	var (
		d    int
		e, f int = 30, 40
		g, h     = 50, 60
	)

	external.Printf("d is = %d\n", d) // d is = 0
	external.Printf("e is = %d\n", e) // e is = 30
	external.Printf("f is = %d\n", f) // f is = 40
	external.Printf("g is = %d\n", g) // g is = 50
	external.Printf("h is = %d\n", h) // h is = 60

	foo := 4
	foo = foo + 5
	foo = foo + 6 + 7 + 8
	external.Printf("foo is = %d\n", foo) // foo is = 30

	var a int
	external.Printf("a is = %d\n", a) // a is = 0

	var b int = 20
	external.Printf("b is = %d\n", b) // b is = 20

	var c = 21
	external.Printf("c is = %d\n", c) // c is = 21

	_ = c
}
