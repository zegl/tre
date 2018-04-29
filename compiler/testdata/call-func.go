package main

import "fmt"

func otherfunc() {
	// in other func
	fmt.Println("in other func\n")
}

func main() {
	otherfunc()
}
