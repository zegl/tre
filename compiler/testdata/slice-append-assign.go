package main

import "external"

func main() {
	a := []int64{1, 2, 3}

	// len(3) cap(3) 1 2 3
	 external.Printf("len(%d) cap(%d) %d %d %d\n", len(a), cap(a), a[0], a[1], a[2])


	// len(4) cap(6) 1 2 3 4
	a = append(a, 4)
	external.Printf("len(%d) cap(%d) %d %d %d %d\n", len(a), cap(a), a[0], a[1], a[2], a[3])


	// len(5) cap(6) 1 2 3 4 5000
	a = append(a, 5000)
	external.Printf("len(%d) cap(%d) %d %d %d %d %d\n", len(a), cap(a), a[0], a[1], a[2], a[3], a[4])


	// len(6) cap(12) 1 2 3 4 5000 6000
	b := append(a, 6000)
	external.Printf("len(%d) cap(%d) %d %d %d %d %d %d\n", len(b), cap(b), b[0], b[1], b[2], b[3], b[4], b[5])

	// len(6) cap(6) 1 2 3 4 5000 7000
	a = append(a, 7000)
	external.Printf("len(%d) cap(%d) %d %d %d %d %d %d\n", len(a), cap(a), a[0], a[1], a[2], a[3], a[4], a[5])
}
