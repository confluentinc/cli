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
				Description: "measures environment-selection collisions under concurrent `confluent login` sessions.",
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
											{Command: "login --url http://mock", ExitCode: 0, DurationMs: 12, Stderr: "warning: retrying connection"},
											{Command: "environment use env-595", ExitCode: 0, DurationMs: 8},
										},
									},
									{
										Session: 2, IntendedEnv: "env-596", ObservedEnv: "", Verdict: VerdictError,
										Detail: `"login --url http://mock" failed: exit 1: connection refused`,
										Invocations: []Invocation{
											{Command: "login --url http://mock", ExitCode: 1, DurationMs: 5},
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

func TestSessionRowClassTintsByVerdict(t *testing.T) {
	if got := sessionRowClass(VerdictOK); got != "clean" {
		t.Fatalf("sessionRowClass(ok) = %q, want clean", got)
	}
	for _, v := range []Verdict{VerdictCollision, VerdictCorruption, VerdictError} {
		if got := sessionRowClass(v); got != "dirty" {
			t.Fatalf("sessionRowClass(%s) = %q, want dirty", v, got)
		}
	}
}

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"environment-crosstalk":   "environment-crosstalk",
		"Environment Crosstalk!!": "environment-crosstalk",
		"  leading--trailing  ":   "leading-trailing",
		"multi   space___under":   "multi-space-under",
		"already-lower-case-slug": "already-lower-case-slug",
	}
	for in, want := range cases {
		if got := slugify(in); got != want {
			t.Errorf("slugify(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCodeifyPlainTextIsEscapedUnchanged(t *testing.T) {
	got := string(codeify("A & B < C"))
	if !strings.Contains(got, "&amp;") || !strings.Contains(got, "&lt;") {
		t.Fatalf("codeify did not escape plain text: %q", got)
	}
	if strings.Contains(got, "<code>") {
		t.Fatalf("codeify added <code> with no backticks in input: %q", got)
	}
}

func TestCodeifyRendersBalancedBacktickSpanAsCode(t *testing.T) {
	got := string(codeify("run `environment use` to switch"))
	if !strings.Contains(got, "<code>environment use</code>") {
		t.Fatalf("codeify did not render backtick span as <code>: %q", got)
	}
}

func TestCodeifyEscapesScriptTags(t *testing.T) {
	got := string(codeify("<script>alert(1)</script>"))
	if strings.Contains(got, "<script>") {
		t.Fatalf("codeify emitted unescaped <script>: %q", got)
	}
	if !strings.Contains(got, "&lt;script&gt;") {
		t.Fatalf("codeify did not escape the script tag: %q", got)
	}
}

func TestCodeifyLeavesUnbalancedBacktickLiteral(t *testing.T) {
	got := string(codeify("it`s broken"))
	if strings.Contains(got, "<code>") {
		t.Fatalf("codeify should not emit <code> for an unbalanced backtick: %q", got)
	}
	if !strings.Contains(got, "`") {
		t.Fatalf("codeify should leave the lone backtick literal: %q", got)
	}
}

func TestCleanRuns(t *testing.T) {
	mixed := CellReport{Trials: []TrialResult{{AllPassed: true}, {AllPassed: false}, {AllPassed: true}}}
	if got := cleanRuns(mixed); got != "2 / 3" {
		t.Fatalf("cleanRuns(mixed) = %q, want %q", got, "2 / 3")
	}

	allClean := CellReport{Trials: []TrialResult{{AllPassed: true}, {AllPassed: true}}}
	if got := cleanRuns(allClean); got != "2 / 2" {
		t.Fatalf("cleanRuns(allClean) = %q, want %q", got, "2 / 2")
	}
}

func TestSharedDamage(t *testing.T) {
	mixed := CellReport{Metrics: CellMetrics{CollisionRate: 0.4, CorruptionRate: 0.2, ErrorRate: 0}}
	got := sharedDamage(mixed)
	if !strings.Contains(got, "collision") || !strings.Contains(got, "corruption") {
		t.Fatalf("sharedDamage(mixed) = %q, want mention of collision and corruption", got)
	}
	if strings.Contains(got, "error") {
		t.Fatalf("sharedDamage(mixed) = %q, should not mention error (rate is 0)", got)
	}

	clean := CellReport{Metrics: CellMetrics{}}
	if got := sharedDamage(clean); got != "none" {
		t.Fatalf("sharedDamage(clean) = %q, want %q", got, "none")
	}
}

func TestReportTemplatesParseAndResolveNames(t *testing.T) {
	for _, name := range []string{"index", "scenario", "styles"} {
		if reportTemplates.Lookup(name) == nil {
			t.Fatalf("template %q not found in the parsed embedded set", name)
		}
	}
}

func TestWriteHTMLReportGeneratesIndexAndScenarioPages(t *testing.T) {
	r := sampleReport()
	dir := t.TempDir()

	if err := r.WriteHTMLReport(dir); err != nil {
		t.Fatalf("WriteHTMLReport: %v", err)
	}

	indexData, err := os.ReadFile(filepath.Join(dir, "index.html"))
	if err != nil {
		t.Fatalf("index.html not written: %v", err)
	}
	index := string(indexData)

	scenarioData, err := os.ReadFile(filepath.Join(dir, "environment-crosstalk.html"))
	if err != nil {
		t.Fatalf("environment-crosstalk.html not written: %v", err)
	}
	scenario := string(scenarioData)

	if !strings.Contains(index, `href="environment-crosstalk.html"`) {
		t.Fatalf("index.html does not link to the scenario page: %q", index)
	}
	if !strings.Contains(index, "clean runs") {
		t.Fatalf("index.html missing \"clean runs\" wording: %q", index)
	}

	for _, want := range []string{
		"badge-collision", // drill-down marker carried over from the single-page report
		"All scenarios",   // back-link to index.html
		"<code>",          // codeify rendered a backtick span from the description
	} {
		if !strings.Contains(scenario, want) {
			t.Fatalf("environment-crosstalk.html missing %q", want)
		}
	}

	for _, doc := range []string{index, scenario} {
		if !strings.Contains(doc, "prefers-color-scheme: dark") {
			t.Fatalf("expected page to be theme-aware (missing dark media query): %q", doc)
		}
	}
}
