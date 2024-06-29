package utils

import (
	"testing"

	"github.com/bradleyjkemp/cupaloy/v2"

	"github.com/confluentinc/cli/v3/pkg/flink/test"
)

func TestDefaultPanicRecovery(t *testing.T) {
	stdout := test.RunAndCaptureSTDOUT(t, WithPanicRecovery(func() {
		panic("a panic")
	}))
	cupaloy.SnapshotT(t, stdout)
}

func TestCustomPanicRecovery(t *testing.T) {
	stdout := test.RunAndCaptureSTDOUT(t, WithCustomPanicRecovery(
		func() {
			panic("a panic")
		},
		func() {
			OutputInfo("This is a custom panic recovery")
		}))
	cupaloy.SnapshotT(t, stdout)
}

func TestCustomPanicRecoveryIsSafeWhenCustomRecoveryFuncPanics(t *testing.T) {
	stdout := test.RunAndCaptureSTDOUT(t, WithCustomPanicRecovery(
		func() {
			panic("a panic")
		},
		func() {
			panic("Still safe if we panic here")
		}))
	cupaloy.SnapshotT(t, stdout)
}
