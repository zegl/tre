package main

import "external"

func main() {
	a := [][]int{}
	external.Printf("%d\n", len(a)) // 0

	a = append(a, []int{100})
	external.Printf("%d\n", len(a))  // 1
	external.Printf("%d\n", a[0][0]) // 100

	a = append(a, []int{200, 201})
	external.Printf("%d\n", len(a))  // 2
	external.Printf("%d\n", a[0][0]) // 100
	external.Printf("%d\n", a[1][0]) // 200
	external.Printf("%d\n", a[1][1]) // 201
}
