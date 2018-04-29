package main

import "external"

// 0
// 1
// 2
// 3
// 4
// 5
// 6
// 7
// 8
// 9

func main() {
	for i := 0; i < 10; i = i + 1 {
		external.Printf("%d\n", i)
	}
}
