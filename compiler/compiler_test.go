package compiler

import (
	"errors"
	gobuild "go/build"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"github.com/zegl/tre/cmd/tre/build"
)

func TestAllPrograms(t *testing.T) {
	bindir := gobuild.Default.GOPATH + "/src/github.com/zegl/tre/compiler"
	testsdir := gobuild.Default.GOPATH + "/src/github.com/zegl/tre/compiler/testdata"

	files, _ := ioutil.ReadDir(testsdir)
	if len(files) == 0 {
		t.Error("No test files found")
	}

	for _, file := range files {
		t.Run(file.Name(), func(t *testing.T) {
			if err := buildRunAndCheck(t, bindir, testsdir+"/"+file.Name()); err != nil {
				t.Error("failed: " + err.Error())
			}
		})
	}
}

func buildRunAndCheck(t *testing.T, bindir, path string) error {
	fp, err := os.Stat(path)
	if err != nil {
		return err
	}

	mainPath := path
	if fp.IsDir() {
		mainPath = path + "/main.go"
	}

	content, err := ioutil.ReadFile(mainPath)
	if err != nil {
		return err
	}

	expect := ""

	re, _ := regexp.Compile(`(?m)// (.*?)$`)
	for _, str := range re.FindAllString(string(content), -1) {
		expect += strings.Replace(str, "// ", "", -1) + "\n"
	}

	// Normalize newlines
	expect = strings.Replace(expect, "\r\n", "\n", -1)
	expect = strings.TrimSpace(expect)

	runProgram := true
	var output string

	err = build.Build(path, false)
	if err != nil {
		output = strings.TrimSpace(err.Error())
		runProgram = false
	}

	// Run program output
	if runProgram {
		cmd := exec.Command(bindir + "/output-binary")
		stdout, err := cmd.CombinedOutput()
		if err != nil {
			if err.Error() != "exit status 1" {
				println(path, err.Error())
				t.Log(string(stdout))
				return errors.New("Runtime failure")
			}
		}

		output = output + strings.TrimSpace(string(stdout))
	}

	if expect == output {
		return nil
	}

	t.Logf("Expected:\n---\n'%s'\n---\nResult:\n---\n'%s'\n---\n", expect, output)

	return errors.New("Unexpected result")
}
