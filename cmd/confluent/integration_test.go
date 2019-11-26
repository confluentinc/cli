// +build testrunmain

package main

import (
	"testing"

	"github.com/confluentinc/cli/internal/pkg/test-integ"
)

func TestRunMain(t *testing.T) {
	test_integ.RunTest(t, main)
}
