// +build testrunmain

package main

import (
	"testing"

	"github.com/confluentinc/cli/internal/pkg/test-integ"
)

var (
	argsFilename string
)

func TestRunMain(t *testing.T) {
	isIntegTest = true
	test_integ.RunTest(t, main)
}
