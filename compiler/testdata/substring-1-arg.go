package main

import "external"

func main() {
	mystr := "hello"

	// h
	external.Printf("%c\n", mystr[0])

	// o
	external.Printf("%c\n", mystr[4])

	// runtime panic: index out of range
	external.Printf("%s\n", mystr[5])
}
