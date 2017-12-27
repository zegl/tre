package main

import (
	"github.com/zegl/tre/lexer"
	"github.com/zegl/tre/parser"
	"github.com/zegl/tre/compiler"

	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Printf("No file specified. Usage: %s path/to/file.tre", os.Args[0])
		os.Exit(1)
	}

	inputFile := os.Args[1]
	debug := len(os.Args) > 2 && os.Args[2] == "--debug"

	fileContents, err := ioutil.ReadFile(inputFile)
	if err != nil {
		panic(err)
	}

	lexed := lexer.Lex(string(fileContents))
	parsed := parser.Parse(lexed)

	if debug {
		log.Println(parsed)
	}

	compiled := compiler.Compile(parsed)
	fmt.Println(compiled)
}
