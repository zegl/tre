package main

import "external"

func main() {
	mystr := "hello"

	// hel
	external.Printf("%s\n", mystr[0:3])
	// ell
	external.Printf("%s\n", mystr[1:3])
	// llo
	external.Printf("%s\n", mystr[2:3])
	// lo
	external.Printf("%s\n", mystr[3:3])
	// o
	external.Printf("%s\n", mystr[4:3])
}
