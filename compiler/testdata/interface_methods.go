package main

import "external"

type TestFace interface {
	Foo(int) int
}

type FaceImpl1 struct{}

func (f FaceImpl1) Foo(a int) int {
	external.Printf("FaceImpl1: %d\n", a)
	return 501
}

type FaceImpl2 struct{}

func (f FaceImpl2) Foo(a int) int {
	external.Printf("FaceImpl2: %d\n", a)
	return 1001
}

func main() {
	var face TestFace
	face = FaceImpl1{}
	res1 := face.Foo(500)               // FaceImpl1: 500
	external.Printf("res1: %d\n", res1) // res1: 501

	face = FaceImpl2{}
	res2 := face.Foo(1000)              // FaceImpl2: 1000
	external.Printf("res2: %d\n", res2) // res2: 1001
}
