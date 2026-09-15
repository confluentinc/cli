//go:build eval

package eval

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReportWriteJSONRoundTrips(t *testing.T) {
	r := Report{
		Build: "HEAD",
		Cells: map[string]CellMetrics{
			"shared":   {Trials: 5, CollisionRate: 1.0, PassCaretK: 0.0},
			"isolated": {Trials: 5, CollisionRate: 0.0, PassCaretK: 1.0},
		},
	}
	path := filepath.Join(t.TempDir(), "report.json")

	if err := r.WriteJSON(path); err != nil {
		t.Fatalf("write: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got Report
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Cells["shared"].CollisionRate != 1.0 || got.Cells["isolated"].PassCaretK != 1.0 {
		t.Fatalf("round-trip mismatch: %+v", got)
	}
}

func TestReportSummaryMentionsBothCells(t *testing.T) {
	r := Report{Build: "HEAD", Cells: map[string]CellMetrics{
		"shared": {CollisionRate: 1.0}, "isolated": {CollisionRate: 0.0},
	}}

	s := r.Summary()

	if !strings.Contains(s, "shared") || !strings.Contains(s, "isolated") {
		t.Fatalf("summary missing a cell: %q", s)
	}
}
