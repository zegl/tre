package main

type TestFace interface {
	Foo(int) int
	// Bar(int) int
	//Bar1(int) int
	//Bar2(int) int
	// Baz() string
}

type FaceImpl struct{}

func (f FaceImpl) Foo(a int) int {
	return 123
}

func main() {
	var face TestFace
	face = FaceImpl{}
}
