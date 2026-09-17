//go:build eval

package eval

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type CellReport struct {
	Name    string        `json:"name"` // "shared" | "isolated"
	Metrics CellMetrics   `json:"metrics"`
	Trials  []TrialResult `json:"trials"`
}

type ScenarioReport struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Sessions    int          `json:"sessions"`
	Trials      int          `json:"trials"`
	Cells       []CellReport `json:"cells"`
}

type Report struct {
	Build       string           `json:"build"`
	GeneratedAt string           `json:"generated_at"` // RFC3339
	Scenarios   []ScenarioReport `json:"scenarios"`
}

func (r Report) WriteJSON(path string) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

//go:embed templates/*.tmpl
var templatesFS embed.FS

// reportTemplates is parsed once at package init from the committed template set. A malformed or
// misnamed .tmpl file surfaces immediately (panic on import) rather than silently rendering a blank
// page; TestReportTemplatesParseAndResolveNames additionally guards that "index"/"scenario"/"styles"
// all resolve by name.
var reportTemplates = template.Must(template.New("report").Funcs(template.FuncMap{
	"pct":             formatPercent,
	"rate":            formatRate,
	"cellClean":       cellIsClean,
	"verdictClass":    verdictBadgeClass,
	"tally":           tallyVerdicts,
	"sessionRowClass": sessionRowClass,
	"slugify":         slugify,
	"codeify":         codeify,
	"cleanRuns":       cleanRuns,
	"sharedDamage":    sharedDamage,
	"cellByName":      cellByName,
}).ParseFS(templatesFS, "templates/*.tmpl"))

// scenarioPage is the data a scenario page renders from: the scenario itself plus the report-level
// build metadata that the shared header on every page shows.
type scenarioPage struct {
	ScenarioReport
	Build       string
	GeneratedAt string
}

// WriteHTMLReport renders the multi-page report into dir: an index.html linking to one
// "<slug>.html" per scenario, where slug is the scenario name run through slugify.
func (r Report) WriteHTMLReport(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := reportTemplates.ExecuteTemplate(&buf, "index", r); err != nil {
		return fmt.Errorf("render index: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "index.html"), buf.Bytes(), 0o644); err != nil {
		return err
	}

	for _, sc := range r.Scenarios {
		buf.Reset()
		page := scenarioPage{ScenarioReport: sc, Build: r.Build, GeneratedAt: r.GeneratedAt}
		if err := reportTemplates.ExecuteTemplate(&buf, "scenario", page); err != nil {
			return fmt.Errorf("render scenario %q: %w", sc.Name, err)
		}
		path := filepath.Join(dir, slugify(sc.Name)+".html")
		if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// Summary renders a short human-readable rollup for stdout/test logs.
func (r Report) Summary() string {
	var b strings.Builder
	fmt.Fprintf(&b, "eval report (build %s, generated %s)\n", r.Build, r.GeneratedAt)
	for _, sc := range r.Scenarios {
		fmt.Fprintf(&b, "scenario: %s\n", sc.Name)
		for _, c := range sc.Cells {
			fmt.Fprintf(&b, "  %-9s trials=%d collision_rate=%.3f corruption_rate=%.3f error_rate=%.3f pass^k=%.3f\n",
				c.Name, c.Metrics.Trials, c.Metrics.CollisionRate, c.Metrics.CorruptionRate, c.Metrics.ErrorRate, c.Metrics.PassCaretK)
		}
	}
	return b.String()
}

func formatPercent(f float64) string {
	return fmt.Sprintf("%.1f%%", f*100)
}

func formatRate(f float64) string {
	return fmt.Sprintf("%.3f", f)
}

// cellIsClean reports whether a cell's metrics show no damage at all - the green/red split in the
// HTML summary tables.
func cellIsClean(m CellMetrics) bool {
	return m.CollisionRate == 0 && m.CorruptionRate == 0 && m.ErrorRate == 0 && m.PassCaretK == 1.0
}

// sessionRowClass tints a session's row in the drill-down table so a scan of a collapsed trial's
// table still shows which sessions failed without reading the verdict column.
func sessionRowClass(v Verdict) string {
	if v == VerdictOK {
		return "clean"
	}
	return "dirty"
}

func verdictBadgeClass(v Verdict) string {
	switch v {
	case VerdictOK:
		return "badge-ok"
	case VerdictCollision:
		return "badge-collision"
	case VerdictCorruption:
		return "badge-corruption"
	default:
		return "badge-error"
	}
}

// tallyVerdicts summarizes a trial's sessions as a one-line verdict count for its <summary>.
func tallyVerdicts(sessions []SessionOutcome) string {
	counts := map[Verdict]int{}
	for _, s := range sessions {
		counts[s.Verdict]++
	}
	return fmt.Sprintf("%d ok, %d collision, %d corruption, %d error",
		counts[VerdictOK], counts[VerdictCollision], counts[VerdictCorruption], counts[VerdictError])
}

var slugNonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

// slugify turns a scenario name into a filename-safe slug: lowercase, non-alphanumeric runs
// collapsed to a single "-", with leading/trailing "-" trimmed.
func slugify(s string) string {
	s = strings.ToLower(s)
	s = slugNonAlnum.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

var backtickSpan = regexp.MustCompile("`([^`]+)`")

// codeify HTML-escapes s and then renders balanced backtick-delimited spans as <code>...</code>.
// Escaping runs first so the span contents are already safe, and an unbalanced/leftover backtick is
// left as a literal character rather than eating the rest of the string.
func codeify(s string) template.HTML {
	escaped := template.HTMLEscapeString(s)
	rendered := backtickSpan.ReplaceAllStringFunc(escaped, func(span string) string {
		inner := span[1 : len(span)-1]
		return "<code>" + inner + "</code>"
	})
	return template.HTML(rendered)
}

// cleanRuns reports a cell's trials as "N / total", where N is the count that graded AllPassed
// (i.e. pass^k's numerator) - the plain-language framing the index page leads with.
func cleanRuns(c CellReport) string {
	total := len(c.Trials)
	clean := 0
	for _, t := range c.Trials {
		if t.AllPassed {
			clean++
		}
	}
	return fmt.Sprintf("%d / %d", clean, total)
}

// sharedDamage summarizes a cell's non-zero rates as plain text, e.g. "50.0% collision, 20.0%
// corruption", or "none" if the cell shows no damage at all.
func sharedDamage(c CellReport) string {
	var parts []string
	if c.Metrics.CollisionRate > 0 {
		parts = append(parts, formatPercent(c.Metrics.CollisionRate)+" collision")
	}
	if c.Metrics.CorruptionRate > 0 {
		parts = append(parts, formatPercent(c.Metrics.CorruptionRate)+" corruption")
	}
	if c.Metrics.ErrorRate > 0 {
		parts = append(parts, formatPercent(c.Metrics.ErrorRate)+" error")
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, ", ")
}

// cellByName finds a cell by name, returning the zero value (no panic) if absent. On the index page
// that renders as a "0 / 0" dirty row (cleanRuns -> "0 / 0"; cellClean -> false, since PassCaretK
// 0 != 1.0) rather than a blank one - a known limitation of the current shared/isolated-only index.
// Generalizing the index to render cells by iteration is deferred until the build dimension can
// introduce cells beyond shared/isolated.
func cellByName(cells []CellReport, name string) CellReport {
	for _, c := range cells {
		if c.Name == name {
			return c
		}
	}
	return CellReport{}
}
