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
	face = &FaceImpl{}

	res := face.Add(3)
	external.Printf("res: %d\n", res)

	res = face.Add(4)
	external.Printf("res: %d\n", res)
}
