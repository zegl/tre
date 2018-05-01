package main

import "external"

func foo(bar interface{}) {
	realInt64, ok := bar.(int64)
	if ok {
		external.Printf("is int64: %d\n", realInt64)
	}

	realString, ok := bar.(string)
	if ok {
		external.Printf("is string: %s\n", realString)
	}

	external.Printf("alwaysstring: \"%s\"\n", realString)
}

func main() {
	// is string: foostring
	// alwaysstring: "foostring"
	foo("foostring")

	// is int64: 123
	// alwaysstring: ""
	foo(123)

	// alwaysstring: ""
	foo(false)

	// is string: barstring
	// alwaysstring: "barstring"
	foo("barstring")
}
