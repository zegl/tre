package main

import "external"

type TestFace interface {
	Add(int) int
}

type FaceImpl struct {
	sum int
}

func (f *FaceImpl) Add(val int) int {
	external.Printf("val: %d\n", val)

	f.sum = f.sum + val

	return f.sum
}

func main() {
	var face TestFace

	impl := FaceImpl{}
	face = impl

	// val: 3
	// res: 3
	res := face.Add(3)
	external.Printf("res: %d\n", res)

	// val: 4
	// res: 7
	res = face.Add(4)
	external.Printf("res: %d\n", res)

	// val: 1
	// res: 8
	res = face.Add(1)
	external.Printf("res: %d\n", res)

	// val: 5
	// res: 13
	res = face.Add(5)
	external.Printf("res: %d\n", res)
}
