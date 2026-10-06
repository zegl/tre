package main

import "external"

func counter(start int) func() int {
	return func() int {
		start++
		return start
	}
}

func adder(step int) func(int) int {
	total := 0
	return func(a int) int {
		total = total + a*step
		return total
	}
}

func apply(fn func(int) int, v int) int {
	return fn(v)
}

func main() {
	c1 := counter(100)
	c2 := counter(200)
	external.Printf("%d\n", c1()) // 101
	external.Printf("%d\n", c1()) // 102
	external.Printf("%d\n", c2()) // 201
	external.Printf("%d\n", c1()) // 103

	a := adder(2)
	a(1)
	external.Printf("%d\n", a(10))        // 22
	external.Printf("%d\n", apply(a, 100)) // 222
}
