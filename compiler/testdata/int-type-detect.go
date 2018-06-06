package main

import "external"

func main() {
	var a int8
	a = 100
	external.Printf("%d\n", a) // 100

	var b int16
	b = 200
	external.Printf("%d\n", b) // 200

	var c int32
	c = 300
	external.Printf("%d\n", c) // 300

	var d int64
	d = 400
	external.Printf("%d\n", d) // 400

	var e int
	e = 500
	external.Printf("%d\n", e) // 500
}
