package main

import "external"

func mulDiv(sum int) (x int, y int) {
	x = sum * 2
	y = sum / 2
	return
}


func main() {
	// 34 8
	a, b := mulDiv(17)
	external.Printf("%d %d\n", a, b)
}
