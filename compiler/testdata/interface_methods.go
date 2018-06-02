package main

import "external"

type TestFace interface {
	Foo1(int) int
	Foo2(int)
}

type FaceImpl1 struct{}

func (f FaceImpl1) Foo1(a int) int {
	external.Printf("FaceImpl1: foo1=%d\n", a)
	return 100
}

func (f FaceImpl1) Foo2(a int) {
	external.Printf("FaceImpl1: foo2=%d\n", a)
}

type FaceImpl2 struct{}

func (f FaceImpl2) Foo1(a int) int {
	external.Printf("FaceImpl2: foo1=%d\n", a)
	return 200
}

func (f FaceImpl2) Foo2(a int) {
	external.Printf("FaceImpl2: foo2=%d\n", a)
}

func main() {
	var face TestFace
	face = FaceImpl1{}
	res1 := face.Foo1(500)              // FaceImpl1: foo1=500
	external.Printf("res1: %d\n", res1) // res1: 100
	face.Foo2(505)                      // FaceImpl1: foo2=505

	face = FaceImpl2{}
	res2 := face.Foo1(600)              // FaceImpl2: foo1=600
	external.Printf("res2: %d\n", res2) // res2: 200
	face.Foo2(606)                      // FaceImpl2: foo2=606
}
