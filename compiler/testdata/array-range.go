package main

import "external"


func ff() [2]int{
	var s [2]int
	s[0] = 40
	s[1] = 50
	return s
}

func main() {
	var s [3]int
	s[0] = 10
	s[1] = 20
	s[2] = 30

	// 0 10
	// 1 20
	// 2 30
	for k, v := range s {
		external.Printf("%d %d\n", k, v)
	}

	// 0 40
	// 1 50
	for k, v := range ff() {
		external.Printf("%d %d\n", k, v)
	}
}

