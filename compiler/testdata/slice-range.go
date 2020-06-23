package main

import "external"


func ff() []int {
	return []int{40, 50}
}

func main() {
	s := []int{10, 20, 30}
	// 0 10
	// 1 20
	// 2 30
	for k, v := range s {
		external.Printf("%d %d\n", k, v)
	}

	// _, 10
	// _, 20
	// _, 30
	for _, v := range s {
		external.Printf("_, %d\n", v)
	}

	// 0
	// 1
	// 2
	for k := range s {
	 	external.Printf("%d\n", k)
	}

	// AAA
	// AAA
	// AAA
	for range s {
		external.Printf("%s\n", "AAA")
	}

	// 0 40
	// 1 50
	f := ff()
	for k, v := range f {
		external.Printf("%d %d\n", k, v)
	}

	// 0 40
	// 1 50
	for k, v := range ff() {
		external.Printf("%d %d\n", k, v)
	}
}

