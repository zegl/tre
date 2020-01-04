package main

import (
	"external"
	"sub"
)

func main() {
	external.Printf("%s\n", sub.World()) // World

	var a sub.AnotherInt
	a = 10

	external.Printf("%d\n", a.Plus5()) // 15
}
