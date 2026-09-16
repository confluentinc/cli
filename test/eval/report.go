//go:build eval

package eval

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
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

func (r Report) WriteHTML(path string) error {
	tmpl, err := template.New("report").Funcs(template.FuncMap{
		"pct":             formatPercent,
		"rate":            formatRate,
		"cellClean":       cellIsClean,
		"verdictClass":    verdictBadgeClass,
		"tally":           tallyVerdicts,
		"sessionRowClass": sessionRowClass,
		"trialClean":      trialClean,
	}).Parse(reportHTMLTemplate)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, r); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
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
// HTML summary table.
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

// trialClean reports whether every session in a trial graded ok - the colored rail on the trial's
// collapsed <details> summary.
func trialClean(sessions []SessionOutcome) bool {
	for _, s := range sessions {
		if s.Verdict != VerdictOK {
			return false
		}
	}
	return true
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

const reportHTMLTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>CLI Eval Report</title>
<style>
  body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif; margin: 2rem; color: #1a1a1a; background: #fafafa; }
  h1, h2, h3 { margin: 1.2em 0 0.3em; }
  h1 { margin-top: 0; }
  .meta { color: #666; font-size: 0.9em; margin-bottom: 1.5em; }
  .scenario { border: 1px solid #ddd; border-radius: 6px; padding: 1em 1.5em; margin-bottom: 2em; background: #fff; }
  table { border-collapse: collapse; width: 100%; margin-bottom: 1em; }
  th, td { border: 1px solid #ddd; padding: 6px 10px; text-align: left; font-size: 0.9em; vertical-align: top; }
  th { background: #eee; }
  .scenario table th { position: sticky; top: 0; background: #eee; }
  td.num { text-align: right; font-variant-numeric: tabular-nums; }
  tr.clean td { background: #e9f9ec; }
  tr.dirty td { background: #fdecea; }
  .mono { font-family: SFMono-Regular, Menlo, Consolas, monospace; }
  .badge { display: inline-block; padding: 2px 8px; border-radius: 10px; font-size: 0.78em; font-weight: 600; color: #fff; }
  .badge-ok { background: #1f7a3d; }
  .badge-collision { background: #cc3333; }
  .badge-corruption { background: #a8650a; }
  .badge-error { background: #6b6b6b; }
  details { margin: 0.3em 0; }
  details details { margin-left: 1.25em; }
  details.clean { border-left: 3px solid #2ea44f; padding-left: 0.5em; }
  details.dirty { border-left: 3px solid #cc3333; padding-left: 0.5em; }
  summary { cursor: pointer; font-weight: 600; }
  pre { background: #272822; color: #f8f8f2; padding: 8px 10px; overflow-x: auto; font-size: 0.82em; border-radius: 4px; margin: 0.3em 0; }
  .stream-label { color: #888; font-size: 0.75em; text-transform: uppercase; letter-spacing: 0.04em; margin-top: 0.4em; }
</style>
</head>
<body>
<h1>CLI Eval Report</h1>
<div class="meta">build <span class="mono">{{.Build}}</span> &middot; generated {{.GeneratedAt}}</div>

{{range .Scenarios}}
<div class="scenario">
<h2>{{.Name}}</h2>
<p>{{.Description}}</p>
<table>
<tr><th>Cell</th><th>Trials</th><th>Collision Rate</th><th>Corruption Rate</th><th>Error Rate</th><th>Pass^k</th></tr>
{{range .Cells}}
<tr class="{{if cellClean .Metrics}}clean{{else}}dirty{{end}}">
<td>{{.Name}}</td>
<td class="num">{{.Metrics.Trials}}</td>
<td class="num">{{pct .Metrics.CollisionRate}}</td>
<td class="num">{{pct .Metrics.CorruptionRate}}</td>
<td class="num">{{pct .Metrics.ErrorRate}}</td>
<td class="num">{{rate .Metrics.PassCaretK}}</td>
</tr>
{{end}}
</table>

{{range .Cells}}
<h3>{{.Name}}</h3>
{{range .Trials}}
<details class="{{if trialClean .Sessions}}clean{{else}}dirty{{end}}">
<summary>trial {{.Trial}} &mdash; {{tally .Sessions}}</summary>
<table>
<tr><th>Session</th><th>Intended &rarr; Observed</th><th>Verdict</th><th>Detail</th></tr>
{{range .Sessions}}
<tr class="{{sessionRowClass .Verdict}}">
<td>{{.Session}}</td>
<td class="mono">{{.IntendedEnv}} &rarr; {{if .ObservedEnv}}{{.ObservedEnv}}{{else}}&mdash;{{end}}</td>
<td><span class="badge {{verdictClass .Verdict}}">{{.Verdict}}</span></td>
<td>{{if .Detail}}{{.Detail}}{{else}}&mdash;{{end}}</td>
</tr>
{{end}}
</table>
{{range .Sessions}}
<details>
<summary>session {{.Session}} invocations ({{len .Invocations}})</summary>
{{range .Invocations}}
<div>
<div class="mono">$ {{.Command}} <span style="color:{{if .ExitCode}}#cc3333{{else}}#888{{end}}">[exit {{.ExitCode}}, {{.DurationMs}}ms]</span></div>
{{if .Err}}<pre>err: {{.Err}}</pre>{{end}}
{{if .Stdout}}<div class="stream-label">stdout</div><pre>{{.Stdout}}</pre>{{end}}
{{if .Stderr}}<div class="stream-label">stderr</div><pre>{{.Stderr}}</pre>{{end}}
</div>
{{end}}
</details>
{{end}}
</details>
{{end}}
{{end}}
</div>
{{end}}
</body>
</html>
`
