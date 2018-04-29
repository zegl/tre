package main

import "external"

func variadic(num ...int) int {
	external.Printf("len = %d\n", len(num))
	return 123
}

func variadicWithOtherArgs(preArg int, num ...int) int {
	external.Printf("pre = %d + len = %d\n", preArg, len(num))
	return 123
}

func main() {
	variadic()        // len = 0
	variadic(1)       // len = 1
	variadic(1, 2)    // len = 2
	variadic(1, 2, 3) // len = 3

	variadicWithOtherArgs(100)                // pre = 100 + len = 0
	variadicWithOtherArgs(100, 2, 3)          // pre = 100 + len = 2
	variadicWithOtherArgs(100, 2, 3, 4, 5, 6) // pre = 100 + len = 5

	fromSlice := []int{100, 200, 300}
	variadic(fromSlice...) // len = 3

	variadic([]int{100, 200, 300, 400, 500, 600, 700}...) // len = 7
}
