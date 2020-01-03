package main

import (
	"external"
)

func main() {
	external.Printf("%d\n", 1+2*3)   // 7
	external.Printf("%d\n", 1*2+3)   // 5
	external.Printf("%d\n", (1+2)*3) // 9
	external.Printf("%d\n", 1+(2*3)) // 7
}
