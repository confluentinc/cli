package components

import (
	"testing"

	"github.com/bradleyjkemp/cupaloy/v2"

	"github.com/confluentinc/cli/v4/pkg/flink/test"
)

const maxCol = 100

func TestPrintCompletionsTrue(t *testing.T) {
	actual := test.RunAndCaptureSTDOUT(t, func() {
		PrintCompletionsState(true, maxCol)
	})

	cupaloy.SnapshotT(t, actual)
}

func TestPrintCompletionsFalse(t *testing.T) {
	actual := test.RunAndCaptureSTDOUT(t, func() {
		PrintCompletionsState(false, maxCol)
	})

	cupaloy.SnapshotT(t, actual)
}

func TestPrintDiagnosticsStateTrue(t *testing.T) {
	actual := test.RunAndCaptureSTDOUT(t, func() {
		PrintDiagnosticsState(true, maxCol)
	})

	cupaloy.SnapshotT(t, actual)
}

func TestPrintDiagnosticsStateFalse(t *testing.T) {
	actual := test.RunAndCaptureSTDOUT(t, func() {
		PrintDiagnosticsState(false, maxCol)
	})

	cupaloy.SnapshotT(t, actual)
}
