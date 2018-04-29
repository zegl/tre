package main

import "external"

// 0
// 1
// 2
// 3
// 4

func main() {
	for i := 0; i < 10; i = i + 1 {
		external.Printf("%d\n", i)
		if i == 4 {
			break
		}
	}
}
