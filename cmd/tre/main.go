package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/zegl/tre/compiler/compiler"

	"github.com/zegl/tre/compiler/lexer"
	"github.com/zegl/tre/compiler/parser"

	"io/ioutil"
	"log"
	"os"
)

var debug bool

func main() {
	if len(os.Args) < 2 {
		log.Printf("No file specified. Usage: %s path/to/file.tre", os.Args[0])
		os.Exit(1)
	}

	debug = len(os.Args) > 2 && os.Args[2] == "--debug"

	c := compiler.NewCompiler()

	compilePackage(c, os.Args[1], "main")

	compiled := c.GetIR()

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
		fmt.Println(string(output))
		panic(err)
	}

	if len(output) > 0 {
		fmt.Println(string(output))
		os.Exit(1)
	}

	os.Exit(0)
}

func compilePackage(c *compiler.Compiler, path, name string) {
	f, err := os.Stat(path)
	if err != nil {
		panic(err)
	}

	var parsedFiles []parser.FileNode

	// Parse all files in the folder
	if f.IsDir() {
		files, err := ioutil.ReadDir(path)
		if err != nil {
			panic(path + ": " + err.Error())
		}

		for _, file := range files {
			if !file.IsDir() {
				if strings.HasSuffix(file.Name(), ".go") {
					parsedFiles = append(parsedFiles, parseFile(path+"/"+file.Name()))
				}
			}
		}
	} else {
		// Parse a single file
		parsedFiles = append(parsedFiles, parseFile(path))
	}

	// Scan for ImportNodes
	// Use importNodes to import more packages
	for _, file := range parsedFiles {
		for _, i := range file.Instructions {
			if _, ok := i.(parser.DeclarePackageNode); ok {
				continue
			}

			if importNode, ok := i.(parser.ImportNode); ok {
				compilePackage(c, path+"/vendor/"+importNode.PackagePath, importNode.PackagePath)
				continue
			}

			break
		}
	}

	c.Compile(parser.PackageNode{
		Files: parsedFiles,
		Name:  name,
	})
}

func parseFile(path string) parser.FileNode {
	// Read specified input file
	fileContents, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	// Run input code through the lexer. A list of tokens is returned.
	lexed := lexer.Lex(string(fileContents))

	// Run lexed source through the parser. A syntax tree is returned.
	parsed := parser.Parse(lexed, debug)

	return parsed
}
