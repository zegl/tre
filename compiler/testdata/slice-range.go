package main

import "external"


func main() {
	s := []int{10, 20, 30}
	// 0 10
	// 1 20
	// 2 30
	for k, v := range s {
		external.Printf("%d %d\n", k, v)
	}

	// 0
	// 1
	// 2
	for k := range s {
	 	external.Printf("%d\n", k)
	}

	// AAA
	// AAA
	// AAA
	for range s {
		external.Printf("%s\n", "AAA")
	}
}
