package main

import (
	"go/build"
	"io/ioutil"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func TestAllPrograms(t *testing.T) {
	bindir := build.Default.GOPATH + "/src/github.com/zegl/tre/tests"
	testsdir := build.Default.GOPATH + "/src/github.com/zegl/tre/tests/tests"

	buildOutput, err := exec.Command("go", "build", "-i", "github.com/zegl/tre/cmd/tre").CombinedOutput()
	if err != nil {
		t.Error(err)
		t.Error(string(buildOutput))
		return
	}

	files, _ := ioutil.ReadDir(testsdir)

	if len(files) == 0 {
		t.Error("No test files found")
	}

	for _, file := range files {
		t.Run(file.Name(), func(t *testing.T) {
			if !buildRunAndCheck(t, bindir, testsdir+"/"+file.Name()) {
				t.Error("failed")
			}
		})
	}
}

func buildRunAndCheck(t *testing.T, bindir, path string) bool {
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

	runProgram := true

	// Compile output
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		if err.Error() != "exit status 1" {
			println(path, err.Error())
			t.Log(string(stdout))
			return false
		}

		// Don't execute the program, but check compier message
		runProgram = false
	}

	output := strings.TrimSpace(string(stdout))

	// Run program output
	if runProgram {
		cmd = exec.Command(bindir + "/output-binary")
		stdout, err = cmd.CombinedOutput()
		if err != nil {
			if err.Error() != "exit status 1" {
				println(path, err.Error())
				t.Log(string(stdout))
				return false
			}
		}

		output = output + strings.TrimSpace(string(stdout))
	}

	if expect == output {
		return true
	}

	t.Logf("Expected:\n---\n'%s'\n---\nResult:\n---\n'%s'\n---\n", expect, output)

	return false
}
