// +build testrunmain

package main

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

func init() {
	flag.StringVar(&argsFilename, "args-file", "", "custom args file, newline separated")
	flag.Parse()
}

func printDivider() {
	fmt.Println("END_OF_TEST_OUTPUT")
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

func TestRunMain(t *testing.T) {
	isIntegTest = true
	parsedArgs := []string{}
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
	main()
	var err error
	printDivider()
	if err != nil {
		panic(err)
	}
	if exitCode == 1 {
		t.FailNow()
	}
}
