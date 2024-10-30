package components

import (
	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/confluentinc/cli/v4/pkg/flink/test"
	"testing"
)

const MaxCol = 100

func TestPrintSmartCompletionTrue(t *testing.T) {
	smartCompletion := true

	actual := test.RunAndCaptureSTDOUT(t, func() {
		PrintSmartCompletionState(smartCompletion, MaxCol)
	})

	cupaloy.SnapshotT(t, actual)
}

func TestPrintSmartCompletionFalse(t *testing.T) {
	smartCompletion := false

	actual := test.RunAndCaptureSTDOUT(t, func() {
		PrintSmartCompletionState(smartCompletion, MaxCol)
	})

	cupaloy.SnapshotT(t, actual)
}

func TestPrintSmartDiagnosticsStateTrue(t *testing.T) {
	completions := true

	actual := test.RunAndCaptureSTDOUT(t, func() {
		PrintDiagnosticsState(completions, MaxCol)
	})

	cupaloy.SnapshotT(t, actual)
}

func TestPrintDiagnosticsStateFalse(t *testing.T) {
	completions := false

	actual := test.RunAndCaptureSTDOUT(t, func() {
		PrintDiagnosticsState(completions, MaxCol)
	})

	cupaloy.SnapshotT(t, actual)
}
