package components

import (
	"testing"

	"github.com/bradleyjkemp/cupaloy/v2"

	"github.com/confluentinc/cli/v4/pkg/flink/test"
)

const MaxCol = 100

func TestPrintCompletionsTrue(t *testing.T) {
	completionsEnabled := true

	actual := test.RunAndCaptureSTDOUT(t, func() {
		PrintCompletionsState(completionsEnabled, MaxCol)
	})

	cupaloy.SnapshotT(t, actual)
}

func TestPrintCompletionsFalse(t *testing.T) {
	completionsEnabled := false

	actual := test.RunAndCaptureSTDOUT(t, func() {
		PrintCompletionsState(completionsEnabled, MaxCol)
	})

	cupaloy.SnapshotT(t, actual)
}

func TestPrintDiagnosticsStateTrue(t *testing.T) {
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
