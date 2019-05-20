package main

import "external"

type f struct {
	a int64
}

func main() {
	// 56
	external.Printf("%d\n",  100 / 3 / 4 * 7)

	// 56
	f1 := f {a : 3}
	external.Printf("%d\n",  100 / f1.a / 4 * 7)

	// 10
	external.Printf("%d\n",  2 * 3 + 4)

	// 14
	external.Printf("%d\n",  2 + 3 * 4)
}