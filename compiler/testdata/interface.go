package main

func foo(bar interface{}) {
}

func main() {
	foo(123)
	foo(false)
	foo("foobar")
}
