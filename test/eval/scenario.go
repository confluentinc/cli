//go:build eval

package eval

import (
	"fmt"
	"os"
	"regexp"
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
	Invocations []Invocation // in run order
}

// SessionScript is one concurrent session's ordered workload. Setup steps run to completion first;
// then all sessions release from the barrier together and run their Contend steps, so the contended
// writes overlap. A login step already carries "--url <cloudURL>" (built via loginStep).
type SessionScript struct {
	Label   string
	Setup   []string
	Contend []string
}

// loginStep builds the login command for the mock backend URL, so the "--url" format string is not
// restated across scenario closures.
func loginStep(cloudURL string) string {
	return "login --url " + cloudURL
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
			results[i] = SessionResult{Session: i, HomeDir: home}
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

const (
	envA = "env-596"
	envB = "env-595"
)

// Scenario is one row in the eval matrix: a concurrent workload plus how to grade it.
type Scenario struct {
	Name        string
	Description string
	Sessions    func(cloudURL string) []SessionScript
	Grade       func(results []SessionResult) []SessionOutcome
}

var apiKeyPattern = regexp.MustCompile(`MYKEY[0-9]+`)

// createdGlobalKey extracts the mock-assigned key id this session's `api-key create` printed.
func createdGlobalKey(r SessionResult) string {
	for _, inv := range r.Invocations {
		if m := apiKeyPattern.FindString(inv.Stdout); m != "" {
			return m
		}
	}
	return ""
}

// observeGlobalKeyPresent returns an observe probe for a session's own created global api-key. A key
// that never parsed from stdout is treated as a read error (routes to a visible corruption verdict),
// so a stdout-format change can't silently mask a dropped key as OK.
func observeGlobalKeyPresent(key string) func(SessionResult) (string, error) {
	return func(r SessionResult) (string, error) {
		if key == "" {
			return "", fmt.Errorf("no created api-key parsed from stdout")
		}
		present, err := GlobalAPIKeyPresent(r.HomeDir, key)
		if err != nil {
			return "", err
		}
		if present {
			return key, nil
		}
		return "", nil
	}
}

var crosstalkEnvs = []string{envA, envB}

var Scenarios = []Scenario{
	{
		Name:        "environment-crosstalk",
		Description: "concurrent `confluent login` + `confluent environment use` sessions; shared state lands the wrong active environment, isolated state does not.",
		Sessions: func(cloudURL string) []SessionScript {
			scripts := make([]SessionScript, len(crosstalkEnvs))
			for i, env := range crosstalkEnvs {
				scripts[i] = SessionScript{
					Label:   fmt.Sprintf("uses %s", env),
					Setup:   []string{loginStep(cloudURL)},
					Contend: []string{"environment use " + env},
				}
			}
			return scripts
		},
		Grade: func(results []SessionResult) []SessionOutcome {
			outs := make([]SessionOutcome, len(results))
			for i, r := range results {
				outs[i] = GradeSession(r, crosstalkEnvs[i], func(r SessionResult) (string, error) {
					return ReadActiveEnvironment(r.HomeDir)
				})
			}
			return outs
		},
	},
	{
		Name:        "selection-cross-field",
		Description: "two sessions set DIFFERENT `current_*` fields on one shared context (`kafka cluster use` vs `flink compute-pool use`); a lockless whole-file save drops one.",
		Sessions: func(cloudURL string) []SessionScript {
			return []SessionScript{
				{Label: "uses kafka cluster", Setup: []string{loginStep(cloudURL), "environment use " + envA}, Contend: []string{"kafka cluster use lkc-12345"}},
				{Label: "uses flink compute-pool", Setup: []string{loginStep(cloudURL), "environment use " + envA}, Contend: []string{"flink compute-pool use lfcp-123456"}},
			}
		},
		Grade: func(results []SessionResult) []SessionOutcome {
			return []SessionOutcome{
				GradeSession(results[0], "lkc-12345", func(r SessionResult) (string, error) { return ReadActiveKafkaCluster(r.HomeDir, envA) }),
				GradeSession(results[1], "lfcp-123456", func(r SessionResult) (string, error) { return ReadCurrentFlinkComputePool(r.HomeDir, envA) }),
			}
		},
	},
	{
		Name:        "crud-lost-update",
		Description: "two sessions each `api-key create --resource global` on one shared context; a lockless whole-file save drops one of the two created keys.",
		Sessions: func(cloudURL string) []SessionScript {
			return []SessionScript{
				{Label: "creates key A", Setup: []string{loginStep(cloudURL)}, Contend: []string{"api-key create --resource global"}},
				{Label: "creates key B", Setup: []string{loginStep(cloudURL)}, Contend: []string{"api-key create --resource global"}},
			}
		},
		Grade: func(results []SessionResult) []SessionOutcome {
			outs := make([]SessionOutcome, len(results))
			for i, r := range results {
				key := createdGlobalKey(r)
				outs[i] = GradeSession(r, key, observeGlobalKeyPresent(key))
			}
			return outs
		},
	},
	{
		Name:        "auth-race",
		Description: "one session keeps working (`environment use`) while another `logout`s on the shared context; the logout can clobber the first session's auth, or be resurrected by it.",
		Sessions: func(cloudURL string) []SessionScript {
			return []SessionScript{
				{Label: "stays logged in, selects env", Setup: []string{loginStep(cloudURL)}, Contend: []string{"environment use " + envA}},
				{Label: "logs out", Setup: []string{loginStep(cloudURL)}, Contend: []string{"logout"}},
			}
		},
		Grade: func(results []SessionResult) []SessionOutcome {
			return []SessionOutcome{
				GradeSession(results[0], envA, func(r SessionResult) (string, error) { return ReadActiveEnvironment(r.HomeDir) }),
				GradeSession(results[1], "cleared", func(r SessionResult) (string, error) {
					cleared, err := CredsCleared(r.HomeDir)
					if err != nil {
						return "", err
					}
					if cleared {
						return "cleared", nil
					}
					return "resurrected", nil
				}),
			}
		},
	},
	{
		Name:        "mixed-workload",
		Description: "three sessions doing unrelated real work on one shared context (select an env; create an api-key; use a cluster then log out); under shared state one session's work clobbers another's.",
		Sessions: func(cloudURL string) []SessionScript {
			return []SessionScript{
				{Label: "selects an environment", Setup: []string{loginStep(cloudURL)}, Contend: []string{"environment use " + envA}},
				{Label: "creates an api-key", Setup: []string{loginStep(cloudURL)}, Contend: []string{"api-key create --resource global"}},
				{Label: "uses a cluster then logs out", Setup: []string{loginStep(cloudURL), "environment use " + envA}, Contend: []string{"kafka cluster use lkc-12345", "logout"}},
			}
		},
		Grade: func(results []SessionResult) []SessionOutcome {
			key := createdGlobalKey(results[1])
			return []SessionOutcome{
				GradeSession(results[0], envA, func(r SessionResult) (string, error) { return ReadActiveEnvironment(r.HomeDir) }),
				GradeSession(results[1], key, observeGlobalKeyPresent(key)),
				GradeSession(results[2], "cleared", func(r SessionResult) (string, error) {
					cleared, err := CredsCleared(r.HomeDir)
					if err != nil {
						return "", err
					}
					if cleared {
						return "cleared", nil
					}
					return "resurrected", nil
				}),
			}
		},
	},
}
