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

// SessionScript is one concurrent session's ordered workload. Setup steps run to completion first;
// then all sessions release from the barrier together and run their Contend steps, so the contended
// writes overlap. A login step already carries "--url <cloudURL>".
type SessionScript struct {
	Label   string
	Setup   []string
	Contend []string
}

// RunScenario runs each session's Setup steps, holds every session at a barrier until all have
// finished setup, then releases them together to run their Contend steps - maximizing the stale-read
// overlap the lost-update scenarios need. A session halts on its first failed step; a session that
// fails in Setup still counts down the barrier (so it never deadlocks) and runs no Contend steps.
// Grading runs on each session's final on-disk state after all sessions finish.
func RunScenario(bin string, p Provisioner, scripts []SessionScript, run CommandFunc) []SessionResult {
	results := make([]SessionResult, len(scripts))

	var setupWG sync.WaitGroup
	setupWG.Add(len(scripts))
	release := make(chan struct{})
	go func() {
		setupWG.Wait()
		close(release)
	}()

	var runWG sync.WaitGroup
	for i, sc := range scripts {
		runWG.Add(1)
		go func(i int, sc SessionScript) {
			defer runWG.Done()
			home := p.HomeDir(i)
			results[i] = SessionResult{Session: i, HomeDir: home, Label: sc.Label}
			env := sessionEnv(home)

			failed := false
			for _, step := range sc.Setup {
				inv := run(bin, env, step)
				results[i].Invocations = append(results[i].Invocations, inv)
				if inv.Failed() {
					failed = true
					break
				}
			}
			setupWG.Done() // finished setup, pass or fail
			if failed {
				return
			}

			<-release // barrier: all sessions run Contend together
			for _, step := range sc.Contend {
				inv := run(bin, env, step)
				results[i].Invocations = append(results[i].Invocations, inv)
				if inv.Failed() {
					break
				}
			}
		}(i, sc)
	}
	runWG.Wait()
	return results
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
