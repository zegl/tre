package main

type MyStruct struct {
	Foo string
	Bar int
}

func main() {
	bar := MyStruct{
		Foo: "hello bar",
		Bar: 1234,
	}
	printf("%s\n", bar.Foo) // hello bar
	printf("%d\n", bar.Bar) // 1234

	baz := MyStruct{Foo: "hello baz", Bar: 5000}
	printf("%s\n", baz.Foo) // hello baz
	printf("%d\n", baz.Bar) // 5000

	onlyOne := MyStruct{Foo: "hello only one"}
	printf("%s\n", onlyOne.Foo) // hello only one
	printf("%d\n", onlyOne.Bar) // 0
}
