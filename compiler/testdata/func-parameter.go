package main

import "external"
import "fmt"

func add(a int, b int) int {
	fmt.Println("exec add")
	return a + b
}

func addMultiRet(a int, b int) (int, int, int) {
	fmt.Println("exec addMultiRet")
	return a + b, a + b, a + b
}

func addNoRet(a int, b int) {
	fmt.Println("exec addNoRet")
}

func singleret(a int, b int, fn func(int, int) int) int {
	return fn(a, b)
}

func multiret(a int, b int, fn func(int, int) (int, int, int)) int {
	fn(a, b)
	return 200
}

func noret(a int, b int, fn func(int, int)) int {
	fn(a, b)
	return 300
}

func main() {
	// exec add
	// 4
	external.Printf("%d\n", singleret(1, 3, add))


	// exec addMultiRet
	// 200
	external.Printf("%d\n", multiret(1, 3, addMultiRet))

	// exec addNoRet
	// 300
	external.Printf("%d\n", noret(1, 3, addNoRet))
}
