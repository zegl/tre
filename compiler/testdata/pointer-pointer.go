package main

func main() {
	i := 100
	iptr1 := &i
	iptr2 := &iptr1
	// iptr3 := &iptr2

	// 100
	// 100
	// 100
	//external.Printf("%d\n", i)
	//external.Printf("%d\n", *iptr1)
	// external.Printf("%d\n", ***iptr3)
}
