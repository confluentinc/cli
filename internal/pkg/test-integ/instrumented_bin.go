package test_integ

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
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
	fmt.Println("END_OF_TEST_OUTPUT")
}

var (
	ExitCode int
)

func RunTest(t *testing.T, f func()) {
	var parsedArgs []string
	for _, arg := range os.Args {
		if !strings.HasPrefix(arg, "-test.") && !strings.HasPrefix(arg, "-args-file") {
			parsedArgs = append(parsedArgs, arg)
		}
	}
	if len(argsFilename) > 0 {
		customArgs, err := parseCustomArgs()
		parsedArgs = append(parsedArgs, customArgs...)
		if err != nil {
			t.Fatal(err)
		}
	}
	os.Args = parsedArgs
	f()
	printDivider()
	if ExitCode == 1 {
		t.FailNow()
	}
}
