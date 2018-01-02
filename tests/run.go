package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

func main() {
	all := true

	// ./Test /path/to/bin /path/to/tests
	bindir := os.Args[1]
	testsdir := os.Args[2]

	files, _ := ioutil.ReadDir(testsdir)

	for _, file := range files {
		if !test(bindir, testsdir+"/"+file.Name()) {
			all = false
		}
	}

	if all {
		os.Exit(0)
	}

	os.Exit(1)
}

func test(bindir, path string) bool {

	content, _ := ioutil.ReadFile(path)

	expect := ""

	re, _ := regexp.Compile(`(?m)// (.*?)$`)
	for _, str := range re.FindAllString(string(content), -1) {
		expect += strings.Replace(str, "// ", "", -1) + "\n"
	}

	// Normalize newlines
	expect = strings.Replace(expect, "\r\n", "\n", -1)
	expect = strings.TrimSpace(expect)

	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command(bindir+"/tre.exe", path)
	} else {
		cmd = exec.Command(bindir+"/tre", path)
	}

	// Compile output
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		if err.Error() != "exit status 1" {
			println(path, err.Error())
			log.Println(string(stdout))
			return false
		}
	}

	// Run program output
	cmd = exec.Command(bindir + "/output-binary")
	stdout, err = cmd.CombinedOutput()
	if err != nil {
		if err.Error() != "exit status 1" {
			println(path, err.Error())
			log.Println(string(stdout))
			return false
		}
	}

	output := strings.TrimSpace(string(stdout))

	if expect == output {
		fmt.Printf("OK: %s\n", path)
		return true
	}

	fmt.Printf("FAIL: %s\n", path)
	fmt.Printf("Expected:\n---\n'%s'\n---\nResult:\n---\n'%s'\n---\n", expect, output)

	return false
}
