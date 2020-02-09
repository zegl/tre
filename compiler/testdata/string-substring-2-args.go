package main

import "external"

func main() {
	mystr := "hello"

	// hel
	external.Printf("%s\n", mystr[0:3])
	// he
	external.Printf("%s\n", mystr[0:2])
	// l
	external.Printf("%s\n", mystr[2:3])
	// llo
	external.Printf("%s\n", mystr[2:5])
	// empty:
	external.Printf("empty:%s\n", mystr[3:3])
	// ello
	external.Printf("%s\n", mystr[1:5])

	external.Printf("%d\n", len(mystr[0:1])) // 1
	external.Printf("%d\n", len(mystr[0:2])) // 2
	external.Printf("%d\n", len(mystr[0:3])) // 3
	external.Printf("%d\n", len(mystr[0:4])) // 4
	external.Printf("%d\n", len(mystr[2:2])) // 0
	external.Printf("%d\n", len(mystr[2:5])) // 3

	start := 1
	end := 3
	external.Printf("%s\n", mystr[start:end]) // el

	external.Printf("%s\n", mystr[1:]) // ello
	external.Printf("%s\n", mystr[3:]) // lo

	external.Printf("%d\n", len(mystr[1:])) // 4
	external.Printf("%d\n", len(mystr[3:])) // 2
}
