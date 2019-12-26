package main

import (
	"external"
)

func main() {
	var u uint8
	var i int8


	u = 127
	i = 127

	// 127 127
	external.Printf("%hhi %hhu\n", i, u)

	i++
	u++

	// -128 128
	external.Printf("%hhi %hhu\n", i, u)

	i = 0
	u = 0
	i--
	u--

	// -1 255
	external.Printf("%hhi %hhu\n", i, u)
}
