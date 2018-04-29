package main

import "external"

func otherfunc(num int) int {
	// in other func = 100
	external.Printf("in other func = %d\n", num)
	return num + 10
}

func main() {
	res := otherfunc(100)
	// in main func = 110
	external.Printf("in main func = %d\n", res)
}
