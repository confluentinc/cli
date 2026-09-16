//go:build eval

package eval

import (
	"fmt"
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

type Session struct {
	IntendedEnv string
}

type SessionResult struct {
	Session     int
	HomeDir     string
	IntendedEnv string
	Invocations []Invocation // in run order (login, environment use, ...)
}

// RunScenario runs each session concurrently through login -> environment use, holds all sessions at
// a barrier until every one has finished writing its environment selection, then returns. Grading of
// the resulting config happens in the caller. The barrier is phase-2 scaffolding for a future step
// that reads/acts after every session's write has landed; phase 1 has no such post-barrier action, so
// it does not affect results here. Today's deterministic outcome comes from grading each session's
// final on-disk config.json after all writes complete (in shared mode, two sessions writing distinct
// environments to one file leave exactly one intended environment surviving - a genuine clobber).
func RunScenario(bin, cloudURL string, p Provisioner, sessions []Session, run CommandFunc) []SessionResult {
	results := make([]SessionResult, len(sessions))

	var wroteWG sync.WaitGroup // counts down as each session finishes `environment use`
	wroteWG.Add(len(sessions))
	release := make(chan struct{}) // closed once all sessions have written

	// coordinator closes the release channel after everyone has written.
	go func() {
		wroteWG.Wait()
		close(release)
	}()

	var runWG sync.WaitGroup
	for i, s := range sessions {
		runWG.Add(1)
		go func(i int, s Session) {
			defer runWG.Done()
			home := p.HomeDir(i)
			results[i] = SessionResult{Session: i, HomeDir: home, IntendedEnv: s.IntendedEnv}

			env := sessionEnv(home)

			login := run(bin, env, fmt.Sprintf("login --url %s", cloudURL))
			results[i].Invocations = append(results[i].Invocations, login)
			if login.Failed() {
				wroteWG.Done()
				return
			}

			use := run(bin, env, fmt.Sprintf("environment use %s", s.IntendedEnv))
			results[i].Invocations = append(results[i].Invocations, use)
			wroteWG.Done() // signal this session has written (win or lose)
			if use.Failed() {
				return
			}
			<-release // barrier: wait for all sessions to finish writing
		}(i, s)
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
