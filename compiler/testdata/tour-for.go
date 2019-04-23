package main

import (
	"external"
)

func main() {
	sum := 0
	for i := 0; i < 10; i++ {
		sum += i
	}
	// 45
	external.Printf("%d\n", sum)

	sum = 400
	for i := 0; i < 14; i++ {
		sum -= i
	}
	// 309
	external.Printf("%d\n", sum)

	sum = 4
	for i := 1; i < 6; i++ {
		sum *= i
	}
	// 480
	external.Printf("%d\n", sum)
}
