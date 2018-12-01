package main

import "external"

type mytype struct{
	a int
}

func main() {
	var foo *mytype
	// signal: segmentation fault
	foo.a = 100
}