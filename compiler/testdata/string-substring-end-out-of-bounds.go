package main

import "external"

func main() {
	mystr := "hello"
	// runtime panic: substring out of bounds
	external.Printf("%s\n", mystr[1:6])
}
