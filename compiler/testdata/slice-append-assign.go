package main

import "external"

func main() {
	a := []int64{1, 2, 3}

	// a = len(3) cap(3) 1 2 3
	external.Printf("a = len(%d) cap(%d) %d %d %d\n", len(a), cap(a), a[0], a[1], a[2])

	// a = len(4) cap(6) 1 2 3 4
	a = append(a, 4)
	external.Printf("a = len(%d) cap(%d) %d %d %d %d\n", len(a), cap(a), a[0], a[1], a[2], a[3])

	// a = len(5) cap(6) 1 2 3 4 5000
	a = append(a, 5000)
	external.Printf("a = len(%d) cap(%d) %d %d %d %d %d\n", len(a), cap(a), a[0], a[1], a[2], a[3], a[4])

	// b = len(6) cap(12) 1 2 3 4 5000 6000
	b := append(a, 6000)
	external.Printf("b = len(%d) cap(%d) %d %d %d %d %d %d\n", len(b), cap(b), b[0], b[1], b[2], b[3], b[4], b[5])

	// a = len(6) cap(6) 1 2 3 4 5000 7000
	a = append(a, 7000)
	external.Printf("a = len(%d) cap(%d) %d %d %d %d %d %d\n", len(a), cap(a), a[0], a[1], a[2], a[3], a[4], a[5])

	// b = len(7) cap(12) 1 2 3 4 5000 6000 8000
	b = append(b, 8000)
	external.Printf("b = len(%d) cap(%d) %d %d %d %d %d %d %d\n", len(b), cap(b), b[0], b[1], b[2], b[3], b[4], b[5], b[6])
}
