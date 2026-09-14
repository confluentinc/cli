//go:build eval

package eval

import (
	"fmt"
	"os"
	"strings"
	"sync"

	pauth "github.com/confluentinc/cli/v4/pkg/auth"
)

type CommandFunc func(bin string, env []string, args string) error

type Session struct {
	IntendedEnv string
}

type SessionResult struct {
	Session     int
	HomeDir     string
	IntendedEnv string
	RunErr      error
}

// RunScenario runs each session concurrently through login -> environment use, holds all sessions at
// a barrier until every one has finished writing its environment selection, then returns. Grading of
// the resulting config happens in the caller. The barrier makes the shared-state clobber
// deterministic: in shared mode the last writer wins for everyone.
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

			if err := run(bin, env, fmt.Sprintf("login --url %s", cloudURL)); err != nil {
				results[i].RunErr = fmt.Errorf("login: %w", err)
				wroteWG.Done()
				return
			}
			if err := run(bin, env, fmt.Sprintf("environment use %s", s.IntendedEnv)); err != nil {
				results[i].RunErr = fmt.Errorf("environment use: %w", err)
				wroteWG.Done()
				return
			}
			wroteWG.Done() // signal this session has written
			<-release      // barrier: wait for all sessions to finish writing
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
