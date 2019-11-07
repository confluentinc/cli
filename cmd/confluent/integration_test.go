// +build testrunmain

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
)

func init() {
	flag.Bool("help", false, "")
	flag.Bool("v", false, "")
	flag.Bool("version", false, "")
	flag.Parse()
}

func printDivider() {
	fmt.Println("END_OF_TEST_OUTPUT")
}

func TestRunMain(t *testing.T) {
	isIntegTest = true
	parsedArgs := []string{}
	for _, arg := range os.Args {
		if !strings.HasPrefix(arg, "-test.") {
			parsedArgs = append(parsedArgs, arg)
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
