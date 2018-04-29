package main

import "external"

// a
// 1
// bb
// 2
// ccc
// 3

func myfunc(yolo string) int64 {
	external.Printf("%s\n", yolo)
	external.Printf("%d\n", len(yolo))
	return 0
}

func main() {
	myfunc("a")
	myfunc("bb")
	myfunc("ccc")
}
