package main

import "external"

func main() {
	var fs []func() int
	for i := 0; i < 3; i++ {
		fs = append(fs, func() int {
			return i * 10
		})
	}
	for _, f := range fs {
		external.Printf("%d\n", f())
	}
	// 0
	// 10
	// 20

	var gs []func() int
	for _, v := range []int{7, 8, 9} {
		gs = append(gs, func() int {
			return v
		})
	}
	for _, g := range gs {
		external.Printf("%d\n", g())
	}
	// 7
	// 8
	// 9

	for i := 0; i < 10; i++ {
		skip := func() {
			i = i + 2
		}
		skip()
		external.Printf("%d\n", i)
	}
	// 2
	// 5
	// 8
	// 11
}
