package main

import (
	"fmt"
)

func runIntSwitch(a int) {
	switch a {
	case 3:
		fmt.Println("three")
	case 6, 7, 8:
		fmt.Println("six, seven, eight")
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

	runIntSwitch(6) // six, seven, eight
	runIntSwitch(7) // six, seven, eight
	runIntSwitch(8) // six, seven, eight

	runBoolSwitch(false) // false
	runBoolSwitch(true)  // true
}
