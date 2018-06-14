package main

import "external"

func main() {
	a := 100

	if a == 100 {
		external.Printf("%d\n", a) // 100
		a := 200
		external.Printf("%d\n", a) // 200
	}

	external.Printf("%d\n", a) // 100
}
