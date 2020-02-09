package main

import "external"

func multiNamedReturn(inp int) (x int, y int) {
	x = inp * 2
	y = inp / 2
	return
}

func singleNamedReturn(inp int) (res int) {
	res = inp * 3
	return
}

func namedReturnNotUsed(inp int) (res int) {
	res = inp * 3
	return 500
}

func multiNamedReturnRef(inp int) (x int, y int) {
	x = inp * 2
	y = x + 2
	return
}

func main() {
	// 34 8
	a, b := multiNamedReturn(17)
	external.Printf("%d %d\n", a, b)

	// 54
	external.Printf("%d\n", singleNamedReturn(18))

	// 500
	external.Printf("%d\n", namedReturnNotUsed(18))

	// 36 38
	c, d := multiNamedReturnRef(18)
	external.Printf("%d %d\n", c, d)
}
