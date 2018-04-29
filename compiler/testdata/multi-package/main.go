package main  import "external"

import "sub"

// Hello
// World

func main() {
	external.Printf("Hello\n")
	external.Printf("%s\n", sub.World())
}
