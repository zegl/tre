package main

import "external"

func main() {
	a := []int{1, 2}
	external.Printf("%d\n", len(a)) // 2

	b := append(a, 3)
	external.Printf("%d\n", len(a)) // 2
	external.Printf("%d\n", len(b)) // 3

	a = append(a, 4)
	external.Printf("%d\n", len(a)) // 3

	external.Printf("%d %d %d\n", a[0], a[1], a[2]) // 1 2 4
	external.Printf("%d %d %d\n", b[0], b[1], b[2]) // 1 2 3
}
