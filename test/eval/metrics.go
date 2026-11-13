//go:build eval

package eval

import (
	"fmt"
	"strings"

	"github.com/confluentinc/cli/v4/pkg/utils"
)

// Verdict is the outcome of grading one session within a trial.
type Verdict string

const (
	VerdictOK         Verdict = "ok"
	VerdictCollision  Verdict = "collision"
	VerdictCorruption Verdict = "corruption"
	VerdictError      Verdict = "error"
)

type SessionOutcome struct {
	Session     int          `json:"session"`
	Label       string       `json:"label"`
	Intent      string       `json:"intent"`
	Observed    string       `json:"observed"` // "" if never established
	Verdict     Verdict      `json:"verdict"`
	Detail      string       `json:"detail"` // human explanation
	Invocations []Invocation `json:"invocations"`
}

type TrialResult struct {
	Trial     int              `json:"trial"`
	Sessions  []SessionOutcome `json:"sessions"`
	AllPassed bool             `json:"all_passed"`
}

type CellMetrics struct {
	Trials         int     `json:"trials"`
	CollisionRate  float64 `json:"collision_rate"`
	CorruptionRate float64 `json:"corruption_rate"`
	ErrorRate      float64 `json:"error_rate"`
	PassCaretK     float64 `json:"pass_caret_k"`
}

// GradeSession applies the error > corruption > collision > ok ladder to one session. A failed
// invocation always wins; then config integrity; then observe reads the scenario-relevant state and
// the session passes when it equals intent. observe is the only scenario-specific part.
func GradeSession(r SessionResult, intent string, observe func(SessionResult) (string, error)) SessionOutcome {
	outcome := SessionOutcome{Session: r.Session, Label: r.Label, Intent: intent, Invocations: r.Invocations}

	for _, inv := range r.Invocations {
		if inv.Failed() {
			outcome.Verdict = VerdictError
			outcome.Detail = fmt.Sprintf("%q failed: %s", inv.Command, failureReason(inv))
			return outcome
		}
	}
	if err := GradeConfigIntegrity(r.HomeDir); err != nil {
		outcome.Verdict = VerdictCorruption
		outcome.Detail = fmt.Sprintf("config was unparseable or incomplete: %v", err)
		return outcome
	}
	observed, err := observe(r)
	if err != nil {
		outcome.Verdict = VerdictCorruption
		outcome.Detail = fmt.Sprintf("could not read state: %v", err)
		return outcome
	}
	outcome.Observed = observed
	if observed == intent {
		outcome.Verdict = VerdictOK
		return outcome
	}
	outcome.Verdict = VerdictCollision
	outcome.Detail = fmt.Sprintf("observed %q, intended %q (clobbered by a concurrent session)", observed, intent)
	return outcome
}

// failureReason prefers the non-exec Err (timeout, spawn failure); otherwise excerpts stderr.
func failureReason(inv Invocation) string {
	if inv.Err != "" {
		return inv.Err
	}
	return fmt.Sprintf("exit %d: %s", inv.ExitCode, excerpt(inv.Stderr, 200))
}

func excerpt(s string, n int) string {
	return utils.Abbreviate(strings.TrimSpace(s), n)
}

// GradeTrial grades one concurrent trial using the scenario's own per-session grader.
func GradeTrial(trial int, results []SessionResult, grade func([]SessionResult) []SessionOutcome) TrialResult {
	sessions := grade(results)
	allPassed := true
	for _, s := range sessions {
		if s.Verdict != VerdictOK {
			allPassed = false
			break
		}
	}
	return TrialResult{Trial: trial, Sessions: sessions, AllPassed: allPassed}
}

// AssertRedSharedGreenIsolated is every scenario's expectation: shared shows state damage (a
// collision or corruption; an error-only run does not prove crosstalk), isolated is perfectly clean.
func AssertRedSharedGreenIsolated(shared, isolated CellMetrics) []string {
	var v []string
	if shared.CollisionRate == 0 && shared.CorruptionRate == 0 {
		v = append(v, "expected collisions or corruptions under shared state, got none (crosstalk not demonstrated)")
	}
	if isolated.CollisionRate != 0 {
		v = append(v, fmt.Sprintf("expected zero collisions under isolated state, got %.3f", isolated.CollisionRate))
	}
	if isolated.CorruptionRate != 0 {
		v = append(v, fmt.Sprintf("expected zero corruptions under isolated state, got %.3f", isolated.CorruptionRate))
	}
	if isolated.ErrorRate != 0 {
		v = append(v, fmt.Sprintf("expected zero errors under isolated state, got %.3f", isolated.ErrorRate))
	}
	if isolated.PassCaretK != 1.0 {
		v = append(v, fmt.Sprintf("expected isolated pass^k = 1.0, got %.3f", isolated.PassCaretK))
	}
	return v
}

// Aggregate rolls graded trials into per-cell rates.
func Aggregate(trials []TrialResult) CellMetrics {
	m := CellMetrics{Trials: len(trials)}
	if len(trials) == 0 {
		return m
	}

	var totalSessions, collisions, corruptions, errored, passed int
	for _, t := range trials {
		totalSessions += len(t.Sessions)
		for _, s := range t.Sessions {
			switch s.Verdict {
			case VerdictCollision:
				collisions++
			case VerdictCorruption:
				corruptions++
			case VerdictError:
				errored++
			}
		}
		if t.AllPassed {
			passed++
		}
	}
	if totalSessions > 0 {
		m.CollisionRate = float64(collisions) / float64(totalSessions)
		m.CorruptionRate = float64(corruptions) / float64(totalSessions)
		m.ErrorRate = float64(errored) / float64(totalSessions)
	}
	m.PassCaretK = float64(passed) / float64(len(trials))
	return m
}
