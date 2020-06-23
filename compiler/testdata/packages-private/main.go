package main

import (
	"sub"
	"fmt"
)

func main() {
	// compile panic: Can't use private from outside of sub
	fmt.Println(sub.Public())
	fmt.Println(sub.private())
}
