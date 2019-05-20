package main

import (
	"fmt"
)

func runSwitch(a int) {
	switch a {
	case 3:
		fmt.Println("three")
	case 4:
		fmt.Println("four")
		fallthrough
	case 5:
		fmt.Println("five")
	default:
		fmt.Println("default")
	}
}

func main() {
	// five
	a := 5
	runSwitch(a)

	// default
	runSwitch(100)

	// four
	// five
	runSwitch(4)
}
