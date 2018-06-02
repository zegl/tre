package build

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/zegl/tre/compiler/compiler"
	"github.com/zegl/tre/compiler/lexer"
	"github.com/zegl/tre/compiler/parser"
)

var debug bool

func Build(path string, setDebug bool) error {
	c := compiler.NewCompiler()
	debug = setDebug

	err := compilePackage(c, path, "main")
	if err != nil {
		return err
	}

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
		return err
	}

	if len(output) > 0 {
		fmt.Println(string(output))
		return errors.New("Clang failure")
	}

	return nil
}

func compilePackage(c *compiler.Compiler, path, name string) error {
	f, err := os.Stat(path)
	if err != nil {
		return err
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
				// Tre files doesn't have to contain valid Go code, and is used to prevent issues
				// with some of the go tools (like vgo)
				if strings.HasSuffix(file.Name(), ".go") || strings.HasSuffix(file.Name(), ".tre") {
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

			gopath := os.Getenv("HOME") + "/go"
			if gopathFromEnv, ok := os.LookupEnv("GOPATH"); ok {
				gopath = gopathFromEnv
			}

			if importNode, ok := i.(parser.ImportNode); ok {

				// Is built in to the compiler
				if importNode.PackagePath == "external" {
					continue
				}

				searchPaths := []string{
					path + "/vendor/" + importNode.PackagePath,

					// "GOROOT" equivalent
					gopath + "/src/github.com/zegl/tre/pkg/" + importNode.PackagePath,
				}

				importSuccessful := false

				for _, sp := range searchPaths {
					fp, err := os.Stat(sp)
					if err != nil || !fp.IsDir() {
						continue
					}

					if debug {
						log.Printf("Loading %s from %s", importNode.PackagePath, sp)
					}

					err = compilePackage(c, sp, importNode.PackagePath)
					if err != nil {
						return err
					}

					importSuccessful = true
				}

				if !importSuccessful {
					return fmt.Errorf("Unable to import: %s", importNode.PackagePath)
				}

				continue
			}

			break
		}
	}

	return c.Compile(parser.PackageNode{
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
