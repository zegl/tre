package main

import (
	"sub"
	"fmt"
)

func main() {
	// compile panic: Can't use private from outside of sub
	var s1 sub.Public
	var s2 sub.private
}
