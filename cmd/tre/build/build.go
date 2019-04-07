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
	"github.com/zegl/tre/compiler/passes/escape"
)

var debug bool

func Build(path, goroot, outputBinaryPath string, setDebug bool, optimize bool) error {
	c := compiler.NewCompiler()
	debug = setDebug

	err := compilePackage(c, path, goroot, "main")
	if err != nil {
		return err
	}

	compiled := c.GetIR()

	if debug {
		fmt.Println(compiled)
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

	if outputBinaryPath == "" {
		outputBinaryPath = "output-binary"
	}

	clangArgs := []string{
		"-Wno-override-module", // Disable override target triple warnings
		tmpDir+"/main.ll",      // Path to LLVM IR
		"-o", outputBinaryPath, // Output path
	}

	if optimize {
		clangArgs = append(clangArgs, "-O3")
	}

	// Invoke clang compiler to compile LLVM IR to a binary executable
	cmd := exec.Command("clang", clangArgs...)
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

func compilePackage(c *compiler.Compiler, path, goroot, name string) error {
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
			if _, ok := i.(*parser.DeclarePackageNode); ok {
				continue
			}

			if importNode, ok := i.(*parser.ImportNode); ok {

				for _, packagePath := range importNode.PackagePaths {

					// Is built in to the compiler
					if packagePath == "external" {
						continue
					}

					searchPaths := []string{
						path + "/vendor/" + packagePath,
						goroot + "/" + packagePath,
					}

					importSuccessful := false

					for _, sp := range searchPaths {
						fp, err := os.Stat(sp)
						if err != nil || !fp.IsDir() {
							continue
						}

						if debug {
							log.Printf("Loading %s from %s", packagePath, sp)
						}

						err = compilePackage(c, sp, goroot, packagePath)
						if err != nil {
							return err
						}

						importSuccessful = true
					}

					if !importSuccessful {
						return fmt.Errorf("Unable to import: %s", packagePath)
					}
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

	// List of passes to run on the AST
	passes := []func(parser.FileNode) parser.FileNode{
		escape.Escape,
	}
	for _, pass := range passes {
		parsed = pass(parsed)
	}

	return parsed
}
