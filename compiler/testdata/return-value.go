package main

import "external"
import "fmt"

func otherfunc() int {
	// in other func
	fmt.Println("in other func")
	return 100
}

func main() {
	// 100
	external.Printf("%d\n", otherfunc())
}
