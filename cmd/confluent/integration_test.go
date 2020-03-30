// +build testrunmain

package main

import (
	"testing"

	"github.com/confluentinc/bincover"
)

func TestRunMain(t *testing.T) {
	bincover.RunTest(main)
}
