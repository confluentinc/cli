// +build testrunmain

package main

import (
	"flag"
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

func TestRunMain(t *testing.T) {
	isIntegTest = true
	offset := 0
	if isIntegTest {
		for i, arg := range os.Args {
			if strings.Contains(arg, "-test.coverprofile") || strings.Contains(arg, "-test.run") {
				if i+1 < len(os.Args) {
					offset = i
				}
			}
		}
	}
	os.Args = append([]string{os.Args[0]}, os.Args[offset+1:]...)
	main()
	var err error
	os.Stdout, err = os.Open(os.DevNull)
	if err != nil {
		panic(err)
	}
	if exitCode == 1 {
		t.FailNow()
	}
}
