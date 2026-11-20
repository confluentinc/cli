//go:build eval

package eval

import (
	"os"
	"strings"
	"sync"

	pauth "github.com/confluentinc/cli/v4/pkg/auth"
)

// Invocation is one captured `confluent` subprocess run.
type Invocation struct {
	Command    string `json:"command"` // args only, e.g. "environment use env-596" (never the bin path)
	ExitCode   int    `json:"exit_code"`
	Stdout     string `json:"stdout"`
	Stderr     string `json:"stderr"`
	DurationMs int64  `json:"duration_ms"`
	Err        string `json:"err,omitempty"` // non-exec failure (timeout, spawn error)
}

// Failed reports whether the invocation should halt the session (nonzero exit or a non-exec error).
func (i Invocation) Failed() bool {
	return i.ExitCode != 0 || i.Err != ""
}

type CommandFunc func(bin string, env []string, args string) Invocation

type SessionResult struct {
	Session     int
	HomeDir     string
	Label       string       // descriptive role, carried from the script's SessionScript.Label
	Invocations []Invocation // in run order
}

// SessionScript is one concurrent session's ordered workload. Setup steps establish the session's
// starting state; then all sessions release together and run their Contend steps, so the contended
// writes overlap. A login step already carries "--url <cloudURL>".
type SessionScript struct {
	Label   string
	Setup   []string
	Contend []string
}

// RunScenario runs every session's Setup steps one session at a time, then releases all sessions
// together to run their Contend steps. Serial setup keeps shared-home logins and selections from
// racing, so any damage comes from the measured Contend steps; isolated cells follow the same
// schedule so the home directory is the only difference between cells. A session halts on its first
// failed step; one that fails in Setup runs no Contend steps. Grading runs on each session's final
// on-disk state after all sessions finish.
func RunScenario(bin string, p Provisioner, scripts []SessionScript, run CommandFunc) []SessionResult {
	results := make([]SessionResult, len(scripts))
	release := make(chan struct{})
	var wg sync.WaitGroup
	for i, sc := range scripts {
		home := p.HomeDir(i)
		results[i] = SessionResult{Session: i, HomeDir: home, Label: sc.Label}
		env := sessionEnv(home)
		if !runSteps(bin, env, sc.Setup, run, &results[i]) {
			continue
		}
		wg.Add(1)
		go func(i int, sc SessionScript) {
			defer wg.Done()
			<-release // barrier: all sessions run Contend together, after every Setup
			runSteps(bin, env, sc.Contend, run, &results[i])
		}(i, sc)
	}
	close(release)
	wg.Wait()
	return results
}

// runSteps runs steps in order, recording each invocation, and reports whether all of them passed.
func runSteps(bin string, env, steps []string, run CommandFunc, result *SessionResult) bool {
	for _, step := range steps {
		inv := run(bin, env, step)
		result.Invocations = append(result.Invocations, inv)
		if inv.Failed() {
			return false
		}
	}
	return true
}

// sessionEnv builds the per-subprocess environment: the base environment with HOME/USERPROFILE
// overridden to the session's dir and mock credentials injected. Never mutates the global process
// env, so concurrent sessions stay isolated.
func sessionEnv(home string) []string {
	base := os.Environ()
	out := make([]string, 0, len(base)+4)
	for _, kv := range base {
		if strings.HasPrefix(kv, "HOME=") || strings.HasPrefix(kv, "USERPROFILE=") {
			continue
		}
		out = append(out, kv)
	}
	out = append(out,
		// appended last so they win over any inherited values: os/exec keeps the last duplicate key.
		"HOME="+home,
		"USERPROFILE="+home,
		pauth.ConfluentCloudEmail+"=fake@user.com",
		pauth.ConfluentCloudPassword+"=pass1",
	)
	return out
}

// Scenario is one row in the eval matrix: a concurrent workload plus how to grade it.
type Scenario struct {
	Name        string
	Description string
	Sessions    func(cloudURL string) []SessionScript
	Grade       func(results []SessionResult) []SessionOutcome
}
