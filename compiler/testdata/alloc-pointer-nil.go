package main

import "external"

type mytype struct{
	a int
}

func main() {
	var foo *mytype
	// Expected: runtime crash
	foo.a = 100
}