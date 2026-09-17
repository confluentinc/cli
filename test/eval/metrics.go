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
	IntendedEnv string       `json:"intended_env"`
	ObservedEnv string       `json:"observed_env"` // "" if never established
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

// GradeSession grades one session's captured run against the config it left behind, in priority
// order: a failing invocation always wins (never double-counted as a collision), then config
// integrity, then whether the session landed on its intended environment.
func GradeSession(r SessionResult) SessionOutcome {
	outcome := SessionOutcome{
		Session:     r.Session,
		IntendedEnv: r.IntendedEnv,
		Invocations: r.Invocations,
	}

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

	observed, err := ReadActiveEnvironment(r.HomeDir)
	if err != nil {
		outcome.Verdict = VerdictCorruption
		outcome.Detail = fmt.Sprintf("could not read active environment: %v", err)
		return outcome
	}
	outcome.ObservedEnv = observed

	if observed == r.IntendedEnv {
		outcome.Verdict = VerdictOK
		return outcome
	}
	outcome.Verdict = VerdictCollision
	outcome.Detail = fmt.Sprintf("acted on %q, intended %q (clobbered by a concurrent session)", observed, r.IntendedEnv)
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

// GradeTrial grades one concurrent trial's session results into a TrialResult.
func GradeTrial(trial int, results []SessionResult) TrialResult {
	sessions := make([]SessionOutcome, len(results))
	allPassed := true
	for i, r := range results {
		sessions[i] = GradeSession(r)
		if sessions[i].Verdict != VerdictOK {
			allPassed = false
		}
	}
	return TrialResult{Trial: trial, Sessions: sessions, AllPassed: allPassed}
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
