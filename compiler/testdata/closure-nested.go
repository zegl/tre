package main

import "external"

func main() {
	outer := 1
	f := func() func() int {
		mid := 10
		return func() int {
			outer++
			mid++
			return outer + mid
		}
	}

	g := f()
	external.Printf("%d\n", g())     // 13
	external.Printf("%d\n", g())     // 15
	external.Printf("%d\n", outer)   // 3

	h := f()
	external.Printf("%d\n", h())     // 15

	shadow := 100
	k := func(shadow int) int {
		return shadow + outer
	}
	external.Printf("%d\n", k(5))    // 9
	external.Printf("%d\n", shadow)  // 100
}
