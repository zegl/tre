package fmt

import "external"

func Println(a string) {
	external.Printf("%s\n", a)
}

func Printf(format string, a ...interface{}) {
	// This does not work
	// TODO: Figure out how to convert tre-interfaces to vararg C-style calls
	external.Printf(format, a...)
}
