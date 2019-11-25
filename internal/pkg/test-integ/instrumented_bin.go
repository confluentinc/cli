package test_integ

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/bouk/monkey"
)

var (
	argsFilename string
)

const (
	endOfInputDivider = "END_OF_TEST_OUTPUT"
)

func init() {
	flag.StringVar(&argsFilename, "args-file", "", "custom args file, newline separated")
	flag.Parse()
}

func parseCustomArgs() ([]string, error) {
	buf, err := ioutil.ReadFile(argsFilename)
	if err != nil {
		return nil, err
	}
	rawArgs := strings.Split(string(buf), "\n")
	var parsedCustomArgs []string
	for _, arg := range rawArgs {
		arg = strings.TrimSpace(arg)
		if len(arg) > 0 {
			parsedCustomArgs = append(parsedCustomArgs, arg)
		}
	}
	return parsedCustomArgs, nil
}

func printDivider() {
	fmt.Println(endOfInputDivider)
}

func printCoverMode() {
	coverMode := testing.CoverMode()
	if coverMode == "" {
		coverMode = "none"
	}
	fmt.Println(coverMode)
}

func RunTest(t *testing.T, f func()) {
	exitTest := func(code int) {
		t.Fail()
	}
	guard := monkey.Patch(os.Exit, exitTest)
	defer guard.Unpatch()
	var parsedArgs []string
	for _, arg := range os.Args {
		if !strings.HasPrefix(arg, "-test.") && !strings.HasPrefix(arg, "-args-file") {
			parsedArgs = append(parsedArgs, arg)
		}
	}
	if len(argsFilename) > 0 {
		customArgs, err := parseCustomArgs()
		if err != nil {
			t.Fatal(err)
		}
		parsedArgs = append(parsedArgs, customArgs...)
	}
	// Capture stdout. Then format into json?:
	// {output: "blah", "coverMode": "set"}<divider><testOutput>
	// OR output <output><divider><coverMode><testOutput>.
	os.Args = parsedArgs
	f()
	printDivider()
	printCoverMode()
}
