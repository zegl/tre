package main

import "external"

// 0
// 1
// 3
// 4

func main() {
	for i := 0; i < 5; i = i + 1 {
		if i == 2 {
			continue
		}
		external.Printf("%d\n", i)
	}
}
