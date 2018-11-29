package main

import (
	"os"
	"log"

	"github.com/zegl/tre/cmd/tre/build"
)

func main() {
	if len(os.Args) < 2 {
		log.Printf("No file specified. Usage: %s path/to/file.tre", os.Args[0])
		os.Exit(1)
	}

	debug := len(os.Args) > 2 && os.Args[2] == "--debug"

	err := build.Build(os.Args[1], "output-binary", debug)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}
