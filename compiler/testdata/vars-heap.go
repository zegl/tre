package main

import "external"

type Bar struct {
	num int64
}

type Foo struct {
	num int64
	bar *Bar
}

func GetFooPtr() *Foo {
	f := Foo{
		num: 300,
		bar: &Bar{num: 400},
	}

	return &f
}

func main() {
	foo := &Foo{
		num: 100,
		bar: &Bar{num: 200},
	}

	external.Printf("foo.bar.num: %d\n", foo.bar.num) // foo.bar.num: 200
	external.Printf("foo.num: %d\n", foo.num)         // foo.num: 100
	external.Printf("foo.bar.num: %d\n", foo.bar.num) // foo.bar.num: 200
	external.Printf("foo.num: %d\n", foo.num)         // foo.num: 100

	foo2 := GetFooPtr()
	external.Printf("foo2.bar.num: %d\n", foo2.bar.num) // foo2.bar.num: 400
	external.Printf("foo2.num: %d\n", foo2.num)         // foo2.num: 300
	external.Printf("foo2.bar.num: %d\n", foo2.bar.num) // foo2.bar.num: 400
	external.Printf("foo2.num: %d\n", foo2.num)         // foo2.num: 300
}
