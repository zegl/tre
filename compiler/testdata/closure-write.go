package main

import "external"

type point struct {
	x int
	y int
}

func pair() (int, int) {
	return 3, 4
}

func main() {
	count := 0
	inc := func() {
		count++
	}
	inc()
	inc()
	external.Printf("%d\n", count) // 2

	count = 10
	inc()
	external.Printf("%d\n", count) // 11

	pt := point{x: 1, y: 2}
	move := func(dx int) {
		pt.x = pt.x + dx
	}
	move(40)
	external.Printf("%d %d\n", pt.x, pt.y) // 41 2

	a, b := pair()
	swap := func() {
		a, b = b, a
	}
	swap()
	external.Printf("%d %d\n", a, b) // 4 3
}
