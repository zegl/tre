package main

import "external"

func main() {
	f1 := func() int {
		return 100
	}
	external.Printf("%d\n", f1()) // 100

	f2 := func(a int) int {
		return 100 * a
	}
	external.Printf("%d\n", f2(2)) // 200

	func() {
		external.Printf("inside\n") // inside
	}()

	var f3 func(int) int

	f3 = func(a int) int {
		return a + 1
	}

	b := f3(2)

	external.Printf("%d", b)
}
