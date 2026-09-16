//go:build eval

package eval

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func sampleReport() Report {
	return Report{
		Build:       "a1b2c3d",
		GeneratedAt: "2026-09-16T00:00:00Z",
		Scenarios: []ScenarioReport{
			{
				Name:        "environment-crosstalk",
				Description: "measures environment-selection collisions under concurrent CLI sessions.",
				Sessions:    2,
				Trials:      5,
				Cells: []CellReport{
					{
						Name:    "shared",
						Metrics: CellMetrics{Trials: 5, CollisionRate: 0.5, CorruptionRate: 0, ErrorRate: 0, PassCaretK: 0},
						Trials: []TrialResult{
							{
								Trial: 0,
								Sessions: []SessionOutcome{
									{Session: 0, IntendedEnv: "env-596", ObservedEnv: "env-596", Verdict: VerdictOK},
									{
										Session: 1, IntendedEnv: "env-595", ObservedEnv: "env-596", Verdict: VerdictCollision,
										Detail: `acted on "env-596", intended "env-595" (clobbered by a concurrent session)`,
										Invocations: []Invocation{
											{Command: "login --url http://mock", ExitCode: 0, DurationMs: 12},
											{Command: "environment use env-595", ExitCode: 0, DurationMs: 8},
										},
									},
								},
								AllPassed: false,
							},
						},
					},
					{
						Name:    "isolated",
						Metrics: CellMetrics{Trials: 5, CollisionRate: 0, CorruptionRate: 0, ErrorRate: 0, PassCaretK: 1},
						Trials: []TrialResult{
							{
								Trial: 0,
								Sessions: []SessionOutcome{
									{Session: 0, IntendedEnv: "env-596", ObservedEnv: "env-596", Verdict: VerdictOK},
									{Session: 1, IntendedEnv: "env-595", ObservedEnv: "env-595", Verdict: VerdictOK},
								},
								AllPassed: true,
							},
						},
					},
				},
			},
		},
	}
}

func TestReportWriteJSONRoundTrips(t *testing.T) {
	r := sampleReport()
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
	if got.Build != r.Build || got.GeneratedAt != r.GeneratedAt {
		t.Fatalf("round-trip mismatch on header: %+v", got)
	}
	if len(got.Scenarios) != 1 || len(got.Scenarios[0].Cells) != 2 {
		t.Fatalf("round-trip mismatch on scenario/cell shape: %+v", got)
	}
	shared := got.Scenarios[0].Cells[0]
	if shared.Name != "shared" || shared.Metrics.CollisionRate != 0.5 {
		t.Fatalf("round-trip mismatch on shared cell: %+v", shared)
	}
	if len(shared.Trials) != 1 || shared.Trials[0].Sessions[1].Verdict != VerdictCollision {
		t.Fatalf("round-trip mismatch on nested trial/session data: %+v", shared.Trials)
	}
}

func TestReportSummaryMentionsBothCells(t *testing.T) {
	r := sampleReport()

	s := r.Summary()

	if !strings.Contains(s, "shared") || !strings.Contains(s, "isolated") {
		t.Fatalf("summary missing a cell: %q", s)
	}
}

func TestReportWriteHTMLContainsDrillDownMarkers(t *testing.T) {
	r := sampleReport()
	path := filepath.Join(t.TempDir(), "index.html")

	if err := r.WriteHTML(path); err != nil {
		t.Fatalf("write: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	html := string(data)

	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Fatalf("output does not look like HTML")
	}
	for _, want := range []string{
		"shared",
		"isolated",
		"badge-collision",                   // verdict badge class rendered
		"clobbered by a concurrent session", // collision detail substring
		"environment use env-595",           // invocation command
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("HTML missing expected marker %q", want)
		}
	}
}
