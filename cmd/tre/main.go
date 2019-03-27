package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/zegl/tre/cmd/tre/build"
)

func main() {
	if len(os.Args) < 2 {
		log.Printf("No file specified. Usage: %s path/to/file.tre", os.Args[0])
		os.Exit(1)
	}

	fs := flag.NewFlagSet("tre", flag.ExitOnError)
	debug := fs.Bool("debug", false, "Emit debug information during compile time")
	optimize := fs.Bool("optimize", false, "Enable clang optimization")
	err := fs.Parse(os.Args[2:])
	if err != nil {
		panic(err)
	}

	// "GOROOT" (treroot?) detection based on the binary path
	treBinaryPath, _ := os.Executable()
	goroot := filepath.Clean(treBinaryPath + "/../pkg/")

	err = build.Build(os.Args[1], goroot, "output-binary", *debug, *optimize)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}
