package main

import "external"

func main() {
	a := 100

	// a is 100
	if a == 100 {
		external.Printf("a is 100\n")
	} else {
		external.Printf("a is not 100\n")
	}

	// a is not 200
	if a == 200 {
		external.Printf("a is 200\n")
	} else {
		external.Printf("a is not 200\n")
	}

	// a is larger than 50
	if a == 200 {
		external.Printf("a is 200\n")
	} else if a > 300 {
		external.Printf("a is larger than 300\n")
	} else if a > 50 {
		external.Printf("a is larger than 50\n")
	} else {
		external.Printf("a is not 200\n")
	}

	// a is larger than 30
	if a == 200 {
		external.Printf("a is 200\n")
	} else if a > 30 {
		external.Printf("a is larger than 30\n")

		// and a is 100
		if a == 100 {
			external.Printf("and a is 100\n")
		} else {
			external.Printf("and a is not 100\n")
		}

		// and a is not 200
		if a == 200 {
			external.Printf("and a is 200\n")
		} else {
			external.Printf("and a is not 200\n")
		}

	} else if a > 50 {
		external.Printf("a is larger than 50\n")
	} else {
		external.Printf("a is not 200\n")
	}
}
