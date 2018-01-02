package main

import (
	"fmt"
	"os/exec"

	"github.com/zegl/tre/compiler"
	"github.com/zegl/tre/lexer"
	"github.com/zegl/tre/parser"

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

	// Read specified input file
	fileContents, err := ioutil.ReadFile(inputFile)
	if err != nil {
		panic(err)
	}

	// Run input code through the lexer. A list of tokens is returned.
	lexed := lexer.Lex(string(fileContents))

	// Run lexed source through the parser. A syntax tree is returned.
	parsed := parser.Parse(lexed)

	if debug {
		log.Println(parsed)
	}

	// Run AST through the compiler. LLVM IR is returned.
	compiled := compiler.Compile(parsed)

	if debug {
		log.Println(compiled)
	}

	// Get dir to save temporary dirs in
	tmpDir, err := ioutil.TempDir("", "tre")
	if err != nil {
		panic(err)
	}

	// Write LLVM IR to disk
	err = ioutil.WriteFile(tmpDir+"/main.ll", []byte(compiled), 0666)
	if err != nil {
		panic(err)
	}

	// Invoke clang compiler to compile LLVM IR to a binary executable
	cmd := exec.Command("clang",
		tmpDir+"/main.ll",     // Path to LLVM IR
		"-o", "output-binary", // Output path
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}

	if len(output) > 0 {
		fmt.Println(string(output))
		os.Exit(1)
	}

	os.Exit(0)
}
