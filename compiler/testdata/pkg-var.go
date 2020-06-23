package main

import (
	"fmt"
)

type e struct {
	s string
}

var a string = "froop"
var b = "hey"
var c string
var d1, d2 = "d1", "d2"
var e1 = e{s: "fff"}

func main() {
	fmt.Println(a) // froop
	fmt.Println(b) // hey
	c = "bar"
	fmt.Println(c)    // bar
	fmt.Println(d1)   // d1
	fmt.Println(d2)   // d2
	fmt.Println(e1.s) // fff
}
