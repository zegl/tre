package main

import "external"

const (
	a = 10
	c = 12

	big = 100000000

	a1 = iota
	a2 = iota
	a3 = iota
)

const (
	b1 = iota
	b2 = iota
)

func main() {
	external.Printf("%d\n", a) // 10

	var b uint8
	b = 5
	external.Printf("%d\n", b+a) // 15
	external.Printf("%d\n", a+b) // 15

	external.Printf("%d\n", a+c) // 22

	var b32 int32
	b32 = 222288822
	external.Printf("%d\n", b32+a) // 222288832
	external.Printf("%d\n", a+b32) // 222288832

	external.Printf("%d\n", a1) // 3
	external.Printf("%d\n", a2) // 4
	external.Printf("%d\n", a3) // 5

	external.Printf("%d\n", b1) // 0
	external.Printf("%d\n", b2) // 1
}
