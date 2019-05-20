package main

import (
	"fmt"
)

func runIntSwitch(a int) {
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

func runBoolSwitch(a bool) {
	switch a {
	case false:
		fmt.Println("false")
	case true:
		fmt.Println("true")
	}
}

func main() {
	a := 5
	runIntSwitch(a) // five

	runIntSwitch(100) // default

	// four
	// five
	runIntSwitch(4)

	runBoolSwitch(false) // false
	runBoolSwitch(true)  // true
}
