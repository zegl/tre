package main

import "external"

func otherfunc(num int32) int32 {
	if num > int32(50) {
		return int32(500)
	}

	return int32(num)
}

func main() {
	// 20
	external.Printf("%d\n", otherfunc(int32(20)))
	// 500
	external.Printf("%d\n", otherfunc(int32(100)))
}
