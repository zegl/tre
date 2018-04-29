package main

import "external"

func main() {
	var arr [10]int

	for i := 0; i < 10; i = i + 1 {
		arr[i] = i
	}

	slice := arr[2:5]

	// 3
	external.Printf("%d\n", len(slice))
}
