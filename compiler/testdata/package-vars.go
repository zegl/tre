package main

import (
	"external"
)

var foo string
var num int
var sli []int

func main() {
	external.Printf("foo: '%s'\n", foo) // foo: ''
	external.Printf("num: %d\n", num) // num: 0
	external.Printf("sli: %d\n", len(sli)) // sli: 0

	foo = "abc"
	num = 3
	sli = append(sli, 1)
	sli = append(sli, 2)
	sli = append(sli, 3)

	external.Printf("foo: '%s'\n", foo) // foo: 'abc'
	external.Printf("num: %d\n", num) // num: 3
	external.Printf("sli: %d\n", len(sli)) // sli: 3
}
