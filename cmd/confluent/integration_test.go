// +build testrunmain

package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
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
	re := regexp.MustCompile(`^-test\..+`)
	for _, arg := range os.Args {
		if !re.MatchString(arg) {
			parsedArgs = append(parsedArgs, arg)
		}
	}
	os.Args = parsedArgs
	main()
	var err error
	printDivider()
	//os.Stdout, err = os.Open(os.DevNull)
	if err != nil {
		panic(err)
	}
	if exitCode == 1 {
		t.FailNow()
	}
}
