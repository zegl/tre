package main

import "external"

type MyStruct struct {
	Foo string
	Bar int
}

func main() {
	bar := MyStruct{
		Foo: "hello bar",
		Bar: 1234,
	}
	external.Printf("%s\n", bar.Foo) // hello bar
	external.Printf("%d\n", bar.Bar) // 1234

	baz := MyStruct{Foo: "hello baz", Bar: 5000}
	external.Printf("%s\n", baz.Foo) // hello baz
	external.Printf("%d\n", baz.Bar) // 5000

	onlyOne := MyStruct{Foo: "hello only one"}
	external.Printf("%s\n", onlyOne.Foo) // hello only one
	external.Printf("%d\n", onlyOne.Bar) // 0
}
