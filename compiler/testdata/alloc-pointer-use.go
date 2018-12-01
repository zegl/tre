package main

import "external"

type mytype struct{
	a int
}

func main() {
	var foo *mytype
	foo = &mytype{}
	foo.a = 100
	external.Printf("%d\n", foo.a) // 100
}