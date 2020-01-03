package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	flag "github.com/spf13/pflag"
	"github.com/zegl/tre/cmd/tre/build"
)

var (
	debug    bool
	optimize bool
	output   string
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <filename>\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.BoolVarP(&debug, "debug", "d", false, "Emit debug information during compile time")
	flag.BoolVarP(&optimize, "optimize", "O", false, "Enable clang optimization")
	flag.StringVarP(&output, "output", "o", "", "Output binary filename")
}

func main() {
	flag.Parse()

	if len(flag.Args()) < 1 {
		log.Printf("No file specified. Usage: %s path/to/file.tre", os.Args[0])
		os.Exit(1)
	}

	// "GOROOT" (treroot?) detection based on the binary path
	treBinaryPath, _ := os.Executable()

	goroots := []string{
		filepath.Clean(treBinaryPath + "/../pkg/"),
		"/Users/gustav/src/go/src",
	}

	if output == "" {
		basename := filepath.Base(flag.Arg(0))
		output = strings.TrimSuffix(basename, filepath.Ext(basename))
	}

	err := build.Build(flag.Arg(0), goroots, output, debug, optimize)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}
