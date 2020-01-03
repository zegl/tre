package main

import "external"

func main() {
	f1 := func() int {
		return 100
	}

	external.Printf("%d\n", f1()) // 100

	f2 := func(a int) int {
		return 200 + a
	}
	external.Printf("%d\n", f2(2)) // 202

	func() {
		external.Printf("inside\n") // inside
	}()

	var f3 func(int) int

	f3 = func(a int) int {
		return 300 + a
	}

	b := f3(2)
	external.Printf("%d\n", b)     // 302
	external.Printf("%d\n", f3(3)) // 303

	f3 = func(a int) int {
		return 400 + a
	}

	external.Printf("%d\n", f3(5)) // 405
}
