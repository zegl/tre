package main

import "external"

func fib(num int) int {
	if num < 2 {
		return num
	}

	return fib(num-2) + fib(num-1)
}

func main() {
	// fib = 5702887
	external.Printf("fib = %d\n", fib(34))
}
