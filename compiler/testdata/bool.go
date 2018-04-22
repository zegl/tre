package main

func main() {
	// 1
	// 0
	// a was true
	// b was not true

	a := true
	printf("%d\n", a)

	a = false
	printf("%d\n", a)

	a = true
	if a {
		println("a was true")
	}
	if !a {
		println("a was not true")
	}

	b := false
	if b {
		println("b was true")
	}
	if !b {
		println("b was not true"
	}

	var c bool
	c = false
	printf("%d\n", c) // 0
}
