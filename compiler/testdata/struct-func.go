package main

import "external"

type a struct {
	fn func(int) int
}

func main() {
	v := a{
		fn: func(a int) int {
			return a + 1
		},
	}

	// 3
	external.Printf("%d\n", v.fn(2))
}
