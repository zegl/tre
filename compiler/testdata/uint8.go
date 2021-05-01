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

	var d1 uint8
	d1 = 30
	external.Printf("%hhu\n", d1/11) // 2
	external.Printf("%hhu\n", d1/10) // 3
	external.Printf("%hhu\n", d1/9)  // 3
}
