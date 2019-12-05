package test_integ

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"
)

var (
	argsFilename string
	ExitCode     = 0
)

const (
	startOfMetadataMarker = "START_OF_METADATA"
	endOfMetadataMarker   = "END_OF_METADATA"
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

type testMetadata struct {
	CoverMode string `json:"cover_mode"`
	ExitCode  int    `json:"exit_code"`
}

func printMetadata(metadata *testMetadata) {
	fmt.Println(startOfMetadataMarker)
	b, err := json.Marshal(metadata)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
	fmt.Println(endOfMetadataMarker)
}

// RunTest runs function f (usually main), with arguments specified by the flag "args-file", a file of newline-separated args.
// When f runs to completion (success or failure), RunTest prints (newline-separated):
// 1. f's output,
// 2. startOfMetadataMarker
// 3. a testMetadata struct
// 4. endOfMetadataMarker
// It then exits with an exit code of 0.
//
// Otherwise, if an unexpected error is encountered during execution, 
// RunTest prints an error, possibly some additional output, and then exits with an exit code of 1.
func RunTest(t *testing.T, f func()) {
	metadata := new(testMetadata)
	defer printMetadata(metadata)
	var parsedArgs []string
	for _, arg := range os.Args {
		if !strings.HasPrefix(arg, "-test.") && !strings.HasPrefix(arg, "-args-file") {
			parsedArgs = append(parsedArgs, arg)
		}
	}
	if len(argsFilename) > 0 {
		customArgs, err := parseCustomArgs()
		if err != nil {
			log.Fatal(err)
		}
		parsedArgs = append(parsedArgs, customArgs...)
	}
	os.Args = parsedArgs
	f()
	metadata.CoverMode = testing.CoverMode()
	metadata.ExitCode = ExitCode
}
