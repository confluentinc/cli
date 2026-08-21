package agentdetect

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"
	"time"
)

// These tests reconstruct the rows of the failure-mode table in
// agent-detection-signals-comparison.md against a synthetic process tree, so
// the claims in the memo are executable rather than asserted. They need no real
// agent installed and run identically on every platform.

func env(pairs map[string]string) func(string) (string, bool) {
	return func(k string) (string, bool) {
		v, ok := pairs[k]
		return v, ok
	}
}

// notATTY keeps the surface hermetic: without it the TTY signals would differ
// between a developer's terminal and CI.
func notATTY(uintptr) bool { return false }

// envVendorForKey is a test-only mirror of which vendor each env-marker key
// belongs to. Production carries no such mapping for these keys (see
// envFingerprints) since the keys are themselves vendor-specific; this map
// exists purely so scenarios below can assert "claude-code fired" instead of
// spelling out "CLAUDECODE fired". AI_AGENT and AGENT are deliberately
// absent: they indicate agent-ness without naming anyone.
var envVendorForKey = map[string]string{
	"CLAUDECODE":      "claude-code",
	"CLAUDE_CODE":     "claude-code",
	"CURSOR_AGENT":    "cursor",
	"CODEX_SANDBOX":   "codex",
	"CODEX_CI":        "codex",
	"CODEX_THREAD_ID": "codex",
	"GEMINI_CLI":      "gemini-cli",
	"COPILOT_CLI":     "github-copilot",
	"COPILOT_MODEL":   "github-copilot",
	"COPILOT_AGENT":   "github-copilot",
	"REPL_ID":         "replit",
}

// envVendors translates Signals.AgentEnv through envVendorForKey, skipping
// AI_AGENT/AGENT (no vendor) and failing loudly on any key the map doesn't
// cover — which would mean envFingerprints grew a row this test map wasn't
// updated for.
func envVendors(res Result) []string {
	var out []string
	for _, key := range res.Signals.AgentEnv {
		vendor, known := envVendorForKey[key]
		if !known && key != "AI_AGENT" && key != "AGENT" {
			panic("agentdetect emitted env key " + key + " with no entry in envVendorForKey")
		}
		if vendor != "" {
			out = append(out, vendor)
		}
	}
	return out
}

func hasEnvVendor(res Result, vendor string) bool {
	return slices.Contains(envVendors(res), vendor)
}

// tree builds a chain from the caller upward: tree("zsh", "claude") means the
// CLI's parent is zsh and zsh's parent is claude. It is chain() indexed by pid,
// which is the shape a ProcSource has to be.
func tree(names ...string) (fakeSource, int) {
	return sourceFor(chain(names...))
}

func detect(t *testing.T, src fakeSource, start int, vars map[string]string) Result {
	t.Helper()
	return Detect(Options{
		Source: src, StartPid: start, Getenv: env(vars),
		IsTerminal: notATTY, KeepChain: true,
	})
}

// Both signals fire and name the same vendor. They are reported side by side; no
// field in Result claims they agree, because that comparison is analytics' job.
func TestDirectAgentCall(t *testing.T) {
	src, start := tree("zsh", "claude")
	res := detect(t, src, start, map[string]string{"CLAUDECODE": "1"})

	if got := envVendors(res); len(got) != 1 || got[0] != "claude-code" {
		t.Fatalf("env vendors = %v, want [claude-code]", got)
	}
	if res.Signals.AgentAncestor == nil {
		t.Fatal("ancestor = nil, want claude-code")
	}
	if got := res.Signals.AgentAncestor.Vendor; got != "claude-code" {
		t.Errorf("ancestor vendor = %q, want claude-code", got)
	}
	if res.Signals.AgentAncestor.Depth != 2 {
		t.Errorf("depth = %d, want 2", res.Signals.AgentAncestor.Depth)
	}
	if got := res.Signals.AgentAncestor.MatchedOn(); got != "name" {
		t.Errorf("matched_on = %q, want name", got)
	}
	if got := res.Signals.AgentAncestor.Name; got != "claude" {
		t.Errorf("name = %q, want claude", got)
	}
}

// The inheritance false positive — the strongest argument for adding process
// inspection. A human working in a terminal that inherited the variable.
func TestInheritedEnvWithNoAgentAncestor(t *testing.T) {
	src, start := tree("zsh", "tmux", "ghostty")
	res := detect(t, src, start, map[string]string{"CLAUDECODE": "1"})

	// The env var claims an agent; the ancestry contradicts it. Both halves of that
	// contradiction have to survive into the payload for the over-reporting rate to
	// be measurable at all.
	if !hasEnvVendor(res, "claude-code") {
		t.Fatalf("env vendors = %v, want claude-code reported", envVendors(res))
	}
	if res.Signals.AgentAncestor != nil {
		t.Errorf("ancestor = %+v, want nil", res.Signals.AgentAncestor)
	}
}

// The recall gap the name table cannot close: the agent is a node process, so
// the basename says nothing and only argv identifies the vendor.
func TestInterpreterHostedAgentNeedsCmdline(t *testing.T) {
	src := fakeSource{
		1000: {Pid: 1000, Ppid: 1001, Name: "zsh"},
		1001: {Pid: 1001, Ppid: 1, Name: "node", Cmdline: []string{
			"node", "/usr/local/lib/node_modules/@anthropic-ai/claude-code/cli.js",
		}},
	}
	res := detect(t, src, 1000, nil)

	if res.Signals.AgentAncestor == nil {
		t.Fatal("no ancestor matched; argv matching failed to identify a node-hosted agent")
	}
	if got := res.Signals.AgentAncestor.Vendor; got != "claude-code" {
		t.Errorf("vendor = %q, want claude-code", got)
	}
	if got := res.Signals.AgentAncestor.MatchedOn(); got != "argv" {
		t.Errorf("matched_on = %q, want argv", got)
	}
	// The matched pattern is emitted as the evidence. It is a key of our own
	// table, so it carries no user data, and having it in the payload is what
	// lets a bad pattern be found in production data rather than guessed at.
	if got := res.Signals.AgentAncestor.ArgvPattern; got != "@anthropic-ai/claude-code" {
		t.Errorf("argv_pattern = %q, want @anthropic-ai/claude-code", got)
	}
	if got := res.Signals.AgentAncestor.Name; got != "" {
		t.Errorf("name = %q, want empty (node is not an agent name)", got)
	}
	// And with argv unavailable (permission denied, or a platform read that
	// failed for any reason — not Windows-specific) the same tree is a miss.
	blind := fakeSource{
		1000: {Pid: 1000, Ppid: 1001, Name: "zsh"},
		1001: {Pid: 1001, Ppid: 1, Name: "node"},
	}
	if res := detect(t, blind, 1000, nil); res.Signals.AgentAncestor != nil {
		t.Errorf("expected a miss without argv, got %+v", res.Signals.AgentAncestor)
	}
}

// The recall estimator. A candidate host we cannot attribute is counted rather
// than ignored, which is what makes the miss rate measurable from production
// instead of arguable from first principles.
func TestUnattributedInterpreterIsCounted(t *testing.T) {
	src := fakeSource{
		1000: {Pid: 1000, Ppid: 1001, Name: "zsh"},
		1001: {Pid: 1001, Ppid: 1, Name: "node", Cmdline: []string{"node", "/opt/newagent/main.js"}},
	}
	res := detect(t, src, 1000, nil)

	if res.Signals.AgentAncestor != nil {
		t.Fatalf("ancestor = %+v, want nil (no fingerprint for this vendor)", res.Signals.AgentAncestor)
	}
	if len(res.Signals.Unattributed) != 1 {
		t.Fatalf("gaps = %+v, want exactly one", res.Signals.Unattributed)
	}
	g := res.Signals.Unattributed[0]
	if g.Depth != 2 {
		t.Errorf("depth = %d, want 2", g.Depth)
	}
	// Kind is what routes the gap to a fix: interpreter means the argv table is
	// missing a vendor.
	if g.Kind != kindInterpreter {
		t.Errorf("kind = %q, want %q", g.Kind, kindInterpreter)
	}
	// argv WAS readable here, so this is a fingerprint-table gap, not a
	// platform blind spot. The distinction drives different fixes.
	if !g.ArgvReadable {
		t.Error("argv_readable = false, want true (cmdline was present)")
	}
}

// Permission-denied shape (any platform, not just Windows): same tree, no
// argv. Still counted, but flagged as unreachable-by-fingerprint rather than
// as a table gap.
func TestUnreadableArgvIsADistinctGap(t *testing.T) {
	src, start := tree("zsh", "node")
	res := detect(t, src, start, nil)

	if len(res.Signals.Unattributed) != 1 {
		t.Fatalf("gaps = %+v, want exactly one", res.Signals.Unattributed)
	}
	if res.Signals.Unattributed[0].ArgvReadable {
		t.Error("argv_readable = true, want false")
	}
}

// The precision guard that replaced the tier ladder. Argv matching is now
// unrestricted by read level, so the only thing keeping it honest is that it
// does not run against processes whose argv is user-authored text.
//
// This tree is a human typing a command that happens to mention an agent's
// install path. Matching the shell's argv would report an agent call and inflate
// the one number this whole exercise exists to produce.
func TestUserTypedCommandMentioningAnAgentIsNotAMatch(t *testing.T) {
	src := fakeSource{
		1000: {Pid: 1000, Ppid: 1001, Name: "zsh", Cmdline: []string{
			"zsh", "-c", "ls node_modules/@anthropic-ai/claude-code",
		}},
		1001: {Pid: 1001, Ppid: 1, Name: "ghostty"},
	}
	res := detect(t, src, 1000, nil)

	if res.Signals.AgentAncestor != nil {
		t.Errorf("ancestor = %+v, want nil; a shell's argv is user text, not identity",
			res.Signals.AgentAncestor)
	}
	// And it must not be counted as a gap either — a shell is not a candidate
	// agent host, so it belongs in neither the numerator nor the denominator.
	if len(res.Signals.Unattributed) != 0 {
		t.Errorf("unattributed = %+v, want none", res.Signals.Unattributed)
	}
}

// The corroboration case, and the reason evidence is two fields rather than one
// "tier". cursor-agent is in the name table AND names itself in argv; both fire,
// and reporting only the cheaper one would discard a confirmation already paid
// for by the same syscall.
func TestNameAndArgvEvidenceAreBothReported(t *testing.T) {
	src := fakeSource{
		1000: {Pid: 1000, Ppid: 1001, Name: "zsh"},
		1001: {Pid: 1001, Ppid: 1, Name: "cursor-agent", Cmdline: []string{
			"/usr/local/bin/cursor-agent", "--headless",
		}},
	}
	res := detect(t, src, 1000, nil)

	a := res.Signals.AgentAncestor
	if a == nil {
		t.Fatal("ancestor = nil, want cursor")
	}
	if a.Vendor != "cursor" {
		t.Errorf("vendor = %q, want cursor", a.Vendor)
	}
	if a.Name != "cursor-agent" {
		t.Errorf("name = %q, want cursor-agent", a.Name)
	}
	if a.ArgvPattern != "cursor-agent" {
		t.Errorf("argv_pattern = %q, want cursor-agent", a.ArgvPattern)
	}
	if got := a.MatchedOn(); got != "name+argv" {
		t.Errorf("matched_on = %q, want name+argv", got)
	}
}

// A parent that started after its child means the pid was recycled. Walking on
// would fabricate an ancestor — here, a "claude" that never launched anything.
func TestPidReuseStopsWalkBeforeFabricatingAnAncestor(t *testing.T) {
	src := fakeSource{
		1000: {Pid: 1000, Ppid: 1001, Name: "zsh", StartTime: 5000},
		1001: {Pid: 1001, Ppid: 1, Name: "claude", StartTime: 9000},
	}
	res := detect(t, src, 1000, nil)

	if res.Walk.StoppedAt != "pid_reuse" {
		t.Errorf("stopped_at = %q, want pid_reuse", res.Walk.StoppedAt)
	}
	if res.Signals.AgentAncestor != nil {
		t.Errorf("ancestor = %+v, want nil — this agent is not really an ancestor", res.Signals.AgentAncestor)
	}
	if !res.Walk.Truncated {
		t.Error("want truncated = true; the walk did not reach init")
	}
}

// The guard must not fire on ordinary trees, where parents start first, nor on
// platforms that supply no start time at all (StartTime zero).
func TestPidReuseGuardDoesNotFireOnValidChains(t *testing.T) {
	ordered := fakeSource{
		1000: {Pid: 1000, Ppid: 1001, Name: "zsh", StartTime: 9000},
		1001: {Pid: 1001, Ppid: 1, Name: "claude", StartTime: 5000},
	}
	if res := detect(t, ordered, 1000, nil); res.Signals.AgentAncestor == nil {
		t.Error("valid chain: agent ancestor should still be detected")
	}

	// StartTime zero everywhere — the Windows case. Guard disabled, not tripped.
	src, start := tree("zsh", "claude")
	if res := detect(t, src, start, nil); res.Signals.AgentAncestor == nil {
		t.Error("missing start times must disable the guard, not fail the walk")
	}
}

// Env var survives an ssh hop only if forwarded; ancestry never does.
func TestRemoteBoundaryStopsWalk(t *testing.T) {
	src, start := tree("bash", "sshd")
	res := detect(t, src, start, nil)

	if res.Walk.StoppedAt != "remote_boundary" {
		t.Errorf("stopped_at = %q, want remote_boundary", res.Walk.StoppedAt)
	}
	if res.Signals.AgentAncestor != nil {
		t.Errorf("ancestor = %+v, want nil", res.Signals.AgentAncestor)
	}
}

// A task runner between the agent and the CLI must not break detection — this
// is why the walk goes deep instead of checking only the immediate parent.
func TestWrapperChainStillDetects(t *testing.T) {
	src, start := tree("sh", "make", "npm", "bash", "claude")
	res := detect(t, src, start, nil)

	// No env var here: ancestry is the only signal that fires, which is the case
	// the env-var-only approach misses outright.
	if len(envVendors(res)) != 0 {
		t.Fatalf("env vendors = %v, want none", envVendors(res))
	}
	if res.Signals.AgentAncestor == nil {
		t.Fatal("ancestor = nil, want claude-code")
	}
	if res.Signals.AgentAncestor.Depth != 5 {
		t.Errorf("depth = %d, want 5", res.Signals.AgentAncestor.Depth)
	}
}

// Chain composition: the wrappers between an agent and us are themselves the
// finding. `timeout` in the ancestry is evidence for friction log 1 item 3.
func TestWrapperCompositionIsRecorded(t *testing.T) {
	src, start := tree("sh", "timeout", "xargs", "bash", "claude")
	res := detect(t, src, start, nil)

	names := make([]string, 0, len(res.Signals.Wrappers))
	for _, w := range res.Signals.Wrappers {
		names = append(names, w.Name)
	}
	want := []string{"timeout", "xargs"}
	if len(names) != len(want) {
		t.Fatalf("wrappers = %v, want %v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Errorf("wrappers[%d] = %q, want %q", i, names[i], want[i])
		}
	}
	if res.Signals.Wrappers[0].Depth != 2 {
		t.Errorf("timeout depth = %d, want 2", res.Signals.Wrappers[0].Depth)
	}
	// sh → timeout → xargs → bash → claude: two adjacent wrappers, which is the
	// FF-9295 fan-out shape wrapped in a timeout.
	if got := res.Signals.ChainShape; got != "swwsa" {
		t.Errorf("chain_shape = %q, want %q", got, "swwsa")
	}
}

// An argv match reclassifies an interpreter as an agent, and the shape must
// record what it turned out to be, not what it looked like on the way in.
func TestChainShapeReflectsArgvReclassification(t *testing.T) {
	src := fakeSource{
		1000: {Pid: 1000, Ppid: 1001, Name: "zsh"},
		1001: {Pid: 1001, Ppid: 1, Name: "node", Cmdline: []string{
			"node", "/usr/local/lib/node_modules/@anthropic-ai/claude-code/cli.js",
		}},
	}
	res := detect(t, src, 1000, nil)

	if got := res.Signals.ChainShape; got != "sa" {
		t.Errorf("chain_shape = %q, want %q (interpreter matched via argv is an agent)", got, "sa")
	}
	// Unmatched, the same tree stays an interpreter.
	blind := fakeSource{
		1000: {Pid: 1000, Ppid: 1001, Name: "zsh"},
		1001: {Pid: 1001, Ppid: 1, Name: "node"},
	}
	if got := detect(t, blind, 1000, nil).Signals.ChainShape; got != "si" {
		t.Errorf("chain_shape = %q, want %q", got, "si")
	}
}

// An editor ancestor is reported separately and must NOT be counted as an
// agent call — a human in the built-in terminal produces this exact tree.
func TestEditorHostIsNotAnAgent(t *testing.T) {
	src, start := tree("zsh", "cursor")
	res := detect(t, src, start, nil)

	if res.Signals.AgentAncestor != nil {
		t.Errorf("ancestor = %+v, want nil (editor is not an agent)", res.Signals.AgentAncestor)
	}
	if res.Signals.IDEHost == nil || res.Signals.IDEHost.Vendor != "cursor" {
		t.Errorf("ide_host = %+v, want cursor", res.Signals.IDEHost)
	}
}

// Nested agents: the shallowest match wins, because the immediate caller is the
// better attribution. The env var names the outer agent and ancestry the inner
// one, and both survive — the disagreement is evidence, not something to resolve
// here.
func TestNestedAgentsReportShallowest(t *testing.T) {
	src, start := tree("bash", "codex", "bash", "claude")
	res := detect(t, src, start, map[string]string{"CLAUDECODE": "1"})

	if res.Signals.AgentAncestor == nil {
		t.Fatal("ancestor = nil, want codex")
	}
	if got := res.Signals.AgentAncestor.Vendor; got != "codex" {
		t.Errorf("vendor = %q, want codex (shallowest)", got)
	}
	if !hasEnvVendor(res, "claude-code") {
		t.Errorf("env vendors = %v, want claude-code preserved alongside the codex ancestor",
			envVendors(res))
	}
}

// Multiple agent vendors in the environment at once — one agent delegating to
// another's CLI. Every match has to survive: a consumer that reads only the first
// entry would call this a vendor disagreement when the two signals actually agree.
func TestMultipleAgentEnvVendorsAreAllReported(t *testing.T) {
	src, start := tree("bash", "copilot")
	res := detect(t, src, start, map[string]string{"CLAUDECODE": "1", "COPILOT_AGENT": "1"})

	got := envVendors(res)
	if len(got) != 2 {
		t.Fatalf("env vendors = %v, want both claude-code and github-copilot", got)
	}
	for _, want := range []string{"claude-code", "github-copilot"} {
		if !hasEnvVendor(res, want) {
			t.Errorf("env vendors = %v, missing %q", got, want)
		}
	}
	// And the ancestor is one of the two vendors named in the environment, not
	// something to be compared against whichever happened to be listed first.
	if res.Signals.AgentAncestor == nil || res.Signals.AgentAncestor.Vendor != "github-copilot" {
		t.Fatalf("ancestor = %+v, want github-copilot", res.Signals.AgentAncestor)
	}
	if !hasEnvVendor(res, res.Signals.AgentAncestor.Vendor) {
		t.Errorf("ancestor vendor %q should appear in env vendors %v",
			res.Signals.AgentAncestor.Vendor, got)
	}
}

// The proposed-standard variables are REPORTED but attribute nothing: the match
// is in agent_env with the variable named, and no vendor comes out of it. AGENT
// in particular is a common name with unrelated meanings, so a consumer has to be
// able to see it fired and still keep it out of an agent count.
func TestGenericVarIsReportedButAttributesNothing(t *testing.T) {
	src, start := tree("zsh", "ghostty")
	res := detect(t, src, start, map[string]string{"AGENT": "1"})

	if len(res.Signals.AgentEnv) != 1 || res.Signals.AgentEnv[0] != "AGENT" {
		t.Fatalf("agent_env = %v, want [AGENT] — the key is what a payload carries", res.Signals.AgentEnv)
	}
	if got := envVendors(res); len(got) != 0 {
		t.Errorf("env vendors = %v, want none", got)
	}
}

func TestEmptyAndFalsyVarsIgnored(t *testing.T) {
	src, start := tree("zsh")
	res := detect(t, src, start, map[string]string{"CLAUDECODE": "", "CURSOR_AGENT": "false"})

	if len(res.Signals.AgentEnv) != 0 {
		t.Errorf("env = %+v, want none", res.Signals.AgentEnv)
	}
}

func TestCycleGuard(t *testing.T) {
	src := fakeSource{
		1000: {Pid: 1000, Ppid: 1001, Name: "zsh"},
		1001: {Pid: 1001, Ppid: 1000, Name: "bash"},
	}
	res := detect(t, src, 1000, nil)

	if res.Walk.StoppedAt != "cycle" {
		t.Errorf("stopped_at = %q, want cycle", res.Walk.StoppedAt)
	}
}

func TestDepthCapTruncates(t *testing.T) {
	names := make([]string, 30)
	for i := range names {
		names[i] = "bash"
	}
	names[25] = "claude"
	src, start := tree(names...)

	res := Detect(Options{
		Source: src, StartPid: start, Getenv: env(nil),
		IsTerminal: notATTY, MaxDepth: 5,
	})
	if !res.Walk.Truncated {
		t.Error("want truncated walk")
	}
	if res.Signals.AgentAncestor != nil {
		t.Error("agent beyond the depth cap must not be reported")
	}
}

func TestLookupErrorIsNonFatal(t *testing.T) {
	src := fakeSource{1000: {Pid: 1000, Ppid: 4242, Name: "zsh"}} // 4242 missing
	res := detect(t, src, 1000, nil)

	if !res.Walk.LookupFailed {
		t.Errorf("lookup_failed = %v, want true", res.Walk.LookupFailed)
	}
	if res.Walk.StoppedAt != "lookup_error" {
		t.Errorf("stopped_at = %q, want lookup_error", res.Walk.StoppedAt)
	}
}

// slowSource makes every ancestor read cost real time, so the wall-clock budget
// can be exercised without depending on how fast the machine running the test is.
type slowSource struct {
	inner fakeSource
	delay time.Duration
}

func (s slowSource) Info(pid int) (ProcInfo, error) {
	time.Sleep(s.delay)
	return s.inner.Info(pid)
}

// The budget is the guard behind "telemetry can never break an invocation", and it
// is the only one whose absence no other test would notice.
func TestBudgetExpiryStopsWalkAndKeepsPartialResults(t *testing.T) {
	// A deep chain with the agent far enough up that the budget must expire first.
	names := make([]string, 20)
	for i := range names {
		names[i] = "bash"
	}
	names[1] = "claude" // reached early, before the budget runs out
	names[15] = "codex" // never reached
	src, start := tree(names...)

	res := Detect(Options{
		Source:     slowSource{inner: src, delay: 2 * time.Millisecond},
		StartPid:   start,
		Getenv:     env(nil),
		IsTerminal: notATTY,
		Budget:     10 * time.Millisecond,
	})

	if res.Walk.StoppedAt != "budget_exceeded" {
		t.Fatalf("stopped_at = %q, want budget_exceeded", res.Walk.StoppedAt)
	}
	if !res.Walk.Truncated {
		t.Error("truncated = false, want true — the walk did not reach init")
	}
	// The assertion that actually matters: expiring is not failing. Everything
	// reached before the deadline still has to come back, or a slow machine would
	// silently report "no agent" instead of "agent, walk truncated".
	if res.Signals.AgentAncestor == nil {
		t.Fatal("ancestor = nil; ancestors reached before the budget expired must survive")
	}
	if got := res.Signals.AgentAncestor.Vendor; got != "claude-code" {
		t.Errorf("vendor = %q, want claude-code", got)
	}
	if res.Walk.DepthReached < 2 {
		t.Errorf("depth_reached = %d, want at least 2", res.Walk.DepthReached)
	}
	if res.Walk.DepthReached >= 16 {
		t.Errorf("depth_reached = %d; the budget should have stopped the walk well before the second agent",
			res.Walk.DepthReached)
	}
}

// A slow walk must not slow the command down beyond the budget it was given. This
// pins the guard's purpose rather than only its bookkeeping.
func TestDetectStaysWithinItsBudget(t *testing.T) {
	names := make([]string, 20)
	for i := range names {
		names[i] = "bash"
	}
	src, start := tree(names...)

	budget := 20 * time.Millisecond
	began := time.Now()
	res := Detect(Options{
		Source:     slowSource{inner: src, delay: 5 * time.Millisecond},
		StartPid:   start,
		Getenv:     env(nil),
		IsTerminal: notATTY,
		Budget:     budget,
	})
	elapsed := time.Since(began)

	// Budget, plus the one in-flight read that cannot be interrupted, plus slack
	// for scheduling. Deliberately loose: this is a smoke test against an
	// unbounded walk, not a latency benchmark.
	if limit := budget + 50*time.Millisecond; elapsed > limit {
		t.Errorf("Detect took %v, want under %v", elapsed, limit)
	}
	if res.Walk.StoppedAt != "budget_exceeded" {
		t.Errorf("stopped_at = %q, want budget_exceeded", res.Walk.StoppedAt)
	}
}

// Argv is only collected when explicitly asked for. Redaction is a second line of
// defense, not the first: the first is that the argv is not there at all.
func TestArgvIsAbsentUnlessExplicitlyRequested(t *testing.T) {
	src := fakeSource{
		1000: {Pid: 1000, Ppid: 1001, Name: "zsh"},
		1001: {Pid: 1001, Ppid: 1, Name: "node", Cmdline: []string{
			"node", "/opt/lib/@anthropic-ai/claude-code/cli.js", "--password", "hunter2",
		}},
	}

	res := detect(t, src, 1000, nil) // KeepChain, but not ShowCmdlines
	if len(res.Walk.Chain) == 0 {
		t.Fatal("chain is empty; this test needs KeepChain to be on")
	}
	for _, e := range res.Walk.Chain {
		if len(e.Cmdline) != 0 {
			t.Errorf("depth %d: cmdline = %v, want empty without ShowCmdlines", e.Depth, e.Cmdline)
		}
	}

	// With it on, argv appears and the credential in it is scrubbed.
	withArgv := Detect(Options{
		Source: src, StartPid: 1000, Getenv: env(nil), IsTerminal: notATTY,
		KeepChain: true, ShowCmdlines: true,
	})
	var found bool
	for _, e := range withArgv.Walk.Chain {
		for _, arg := range e.Cmdline {
			if arg == "hunter2" {
				t.Error("raw credential survived into the chain")
			}
			if arg == "<redacted>" {
				found = true
			}
		}
	}
	if !found {
		t.Error("no redacted argument found; expected --password's value to be scrubbed")
	}
}

// Chain entries are held to the same vocabulary rule as the payload fields: only
// exact fingerprint-table keys are recorded as names, even though the chain is
// local-only. A diagnostic buffer of arbitrary process names is what later gets
// plumbed somewhere it should not be.
func TestChainRecordsOnlyTableKeyNames(t *testing.T) {
	src := fakeSource{
		1000: {Pid: 1000, Ppid: 1001, Name: "zsh"},
		1001: {Pid: 1001, Ppid: 1002, Name: "my-private-tool"},
		1002: {Pid: 1002, Ppid: 1, Name: "code helper (plugin)"},
	}
	res := detect(t, src, 1000, nil)

	for _, e := range res.Walk.Chain {
		if e.Name == "" {
			continue
		}
		if _, ok := procFingerprints[e.Name]; !ok {
			t.Errorf("chain depth %d recorded name %q, which is not a fingerprint-table key",
				e.Depth, e.Name)
		}
	}
	// Specifically: the unknown binary and the suffix-matched helper contribute a
	// Kind but no name.
	for _, e := range res.Walk.Chain {
		if e.Depth == 2 && e.Name != "" {
			t.Errorf("unknown binary recorded name %q, want empty", e.Name)
		}
		if e.Depth == 3 {
			if e.Name != "" {
				t.Errorf("suffix-matched helper recorded name %q, want empty", e.Name)
			}
			if e.Kind != kindIDEHost {
				t.Errorf("depth 3 kind = %q, want %q", e.Kind, kindIDEHost)
			}
		}
	}
}

// A lookup failure must not carry the error text, which on Linux embeds
// /proc/<pid> paths — an observation, not vocabulary.
func TestLookupErrorDoesNotRecordErrorText(t *testing.T) {
	src := fakeSource{1000: {Pid: 1000, Ppid: 4242, Name: "zsh"}}
	res := detect(t, src, 1000, nil)

	for _, e := range res.Walk.Chain {
		if strings.Contains(e.Error, "no such process") {
			t.Errorf("chain depth %d carries the raw error %q", e.Depth, e.Error)
		}
	}
}

// The false positive that arrives through the unknown door. Every binary missing
// from the name table is kindUnknown and therefore argv-eligible, so without a
// position restriction an ordinary command that MENTIONS an agent path gets
// attributed as one — the same error TestUserTypedCommandMentioningAnAgentIsNotAMatch
// prevents for shells, landing in the same headline number.
func TestAgentPathInAnArgumentIsNotAMatch(t *testing.T) {
	cases := map[string][]string{
		"ripgrep over an install path": {"rg", "@anthropic-ai/claude-code"},
		"docker mounting an agent dir": {"docker", "run", "-v", "/opt/aider/:/x", "alpine"},
		"tar of an agent directory":    {"tar", "czf", "backup.tgz", "site-packages/aider"},
	}
	for name, argv := range cases {
		t.Run(name, func(t *testing.T) {
			src := fakeSource{
				1000: {Pid: 1000, Ppid: 1001, Name: "zsh"},
				1001: {Pid: 1001, Ppid: 1, Name: argv[0], Cmdline: argv},
			}
			res := detect(t, src, 1000, nil)

			if res.Signals.AgentAncestor != nil {
				t.Errorf("ancestor = %+v, want nil; an argument is not the program's identity",
					res.Signals.AgentAncestor)
			}
			// Still counted as an unattributed unknown — it IS a process we could
			// not type, and hiding that would understate the blind spot.
			if len(res.Signals.Unattributed) != 1 {
				t.Errorf("unattributed = %+v, want exactly one", res.Signals.Unattributed)
			}
		})
	}
}

// The other half of the same rule: argv[0] still identifies an agent invoked by
// its own install path under a name we have no fingerprint for.
func TestAgentPathInArgvZeroStillMatches(t *testing.T) {
	src := fakeSource{
		1000: {Pid: 1000, Ppid: 1001, Name: "zsh"},
		1001: {Pid: 1001, Ppid: 1, Name: "2.1.219", Cmdline: []string{
			"/opt/node_modules/@anthropic-ai/claude-code/bin/wrapper", "--print",
		}},
	}
	res := detect(t, src, 1000, nil)

	if res.Signals.AgentAncestor == nil || res.Signals.AgentAncestor.Vendor != "claude-code" {
		t.Fatalf("ancestor = %+v, want claude-code from argv[0]", res.Signals.AgentAncestor)
	}
}

// Interpreters name what they are running a couple of words in, and flags in
// between must not stop the scan.
func TestInterpreterIdentityArgumentsAreScannedPastFlags(t *testing.T) {
	cases := map[string][]string{
		"flag before the script": {"node", "--experimental-vm-modules", "/opt/@openai/codex/bin/x.js"},
		"launcher subcommands":   {"uv", "tool", "run", "site-packages/aider"},
	}
	for name, argv := range cases {
		t.Run(name, func(t *testing.T) {
			src := fakeSource{
				1000: {Pid: 1000, Ppid: 1001, Name: "zsh"},
				1001: {Pid: 1001, Ppid: 1, Name: argv[0], Cmdline: argv},
			}
			if res := detect(t, src, 1000, nil); res.Signals.AgentAncestor == nil {
				t.Errorf("ancestor = nil, want a match for %v", argv)
			}
		})
	}

	// But not arbitrarily far in: past the identity positions, arguments are the
	// program's input.
	deep := fakeSource{
		1000: {Pid: 1000, Ppid: 1001, Name: "zsh"},
		1001: {Pid: 1001, Ppid: 1, Name: "node", Cmdline: []string{
			"node", "build.js", "clean", "bundle", "@anthropic-ai/claude-code",
		}},
	}
	if res := detect(t, deep, 1000, nil); res.Signals.AgentAncestor != nil {
		t.Errorf("ancestor = %+v, want nil; that path is an argument, not the script",
			res.Signals.AgentAncestor)
	}
}

// CI reports normalized provider ids, not the variable names they came from, and
// the bare CI variable is kept separate because it co-occurs with all of them.
func TestCIReportsNormalizedProviders(t *testing.T) {
	src, start := tree("bash", "runner")

	res := detect(t, src, start, map[string]string{"CI": "true", "GITHUB_ACTIONS": "true"})
	if got := res.Signals.CI; len(got) != 1 || got[0] != "github-actions" {
		t.Errorf("ci = %v, want [github-actions]", got)
	}
	if res.Signals.CIGeneric {
		t.Error("ci_generic = true, want false when a provider was identified")
	}

	// A CI system we have no variable for: the bare flag is all there is, and it
	// still has to be reported.
	bare := detect(t, src, start, map[string]string{"CI": "1"})
	if len(bare.Signals.CI) != 0 {
		t.Errorf("ci = %v, want none", bare.Signals.CI)
	}
	if !bare.Signals.CIGeneric {
		t.Error("ci_generic = false, want true for a bare CI variable")
	}

	// And no CI at all reports an empty list rather than a nil one, so "not in CI"
	// and "we did not look" stay distinguishable.
	none := detect(t, src, start, nil)
	if none.Signals.CI == nil {
		t.Error("ci = nil, want an empty list")
	}
	if none.Signals.CIGeneric {
		t.Error("ci_generic = true, want false")
	}
}

// The security commitment, enforced rather than asserted in a comment: raw process
// names, pids and argv cannot be serialized out of a Result, whatever a caller
// does with it. This is the test that fails if someone gives ChainEntry JSON tags
// or drops the json:"-" on WalkMeta.Chain.
func TestResultSerializationCannotCarryRawObservations(t *testing.T) {
	src := fakeSource{
		1000: {Pid: 1000, Ppid: 1001, Name: "zsh", Cmdline: []string{"zsh"}},
		1001: {Pid: 1001, Ppid: 1, Name: "my-secret-internal-tool", Cmdline: []string{
			"my-secret-internal-tool", "--password", "hunter2", "/home/someone/private/path",
		}},
	}
	res := Detect(Options{
		Source: src, StartPid: 1000, Getenv: env(nil), IsTerminal: notATTY,
		KeepChain: true, ShowCmdlines: true, // the most permissive settings there are
	})
	if len(res.Walk.Chain) == 0 {
		t.Fatal("chain is empty; this test needs the chain populated to be meaningful")
	}

	encoded, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	payload := string(encoded)

	for _, forbidden := range []string{
		"my-secret-internal-tool", // an observed process name
		"/home/someone/private",   // an argv path
		"hunter2",                 // a credential
		`"chain":`,                // the diagnostic buffer itself
		"1001",                    // a pid
	} {
		if strings.Contains(payload, forbidden) {
			t.Errorf("serialized Result contains %q:\n%s", forbidden, payload)
		}
	}
}

// Slices are always non-nil once detection has run: absent must mean "this CLI
// could not tell us", never "we looked and found nothing".
func TestEmptySignalsAreEmptyNotNil(t *testing.T) {
	src, start := tree("zsh", "ghostty")
	res := detect(t, src, start, nil)

	if res.Signals.AgentEnv == nil {
		t.Error("agent_env = nil, want empty")
	}
	if res.Signals.Unattributed == nil {
		t.Error("unattributed = nil, want empty")
	}
	if res.Signals.Wrappers == nil {
		t.Error("wrappers = nil, want empty")
	}
	if res.Signals.CI == nil {
		t.Error("ci = nil, want empty")
	}
}

func TestRedactionCoversCredentialsInArgv(t *testing.T) {
	got := redact([]string{
		"confluent", "login", "--password", "hunter2",
		"--url", "https://example.com",
		"KAFKA_API_SECRET=abc123",
		"PATH=/usr/bin",
	})
	want := []string{
		"confluent", "login", "--password", "<redacted>",
		"--url", "https://example.com",
		"KAFKA_API_SECRET=<redacted>",
		"PATH=/usr/bin",
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("argv[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// Observed on darwin/arm64, 2026-07-28: Claude Code's native install names the
// binary by version (~/.local/share/claude/versions/2.1.219), so the cheap
// kernel name field reads "2.1.219". No fingerprint table can match that, and
// no amount of table maintenance fixes it. The usable name exists only in the
// executable path, which is why the truncated kernel name is not enough on its
// own — and why, when the path is also unhelpful, argv is the only way through.
func TestVersionNamedBinaryDefeatsCommMatching(t *testing.T) {
	src, start := tree("zsh", "2.1.219")
	res := detect(t, src, start, nil)

	if res.Signals.AgentAncestor != nil {
		t.Fatalf("ancestor = %+v; a version-named binary cannot be matched by name", res.Signals.AgentAncestor)
	}
	// And with the exec path resolved, the same process matches by name.
	src2, start2 := tree("zsh", "claude")
	if res := detect(t, src2, start2, nil); res.Signals.AgentAncestor == nil {
		t.Error("exec-path name should match the name table")
	}
}

// scenarioExpectation pins down the outcome each entry in scenarios_test.go is
// supposed to demonstrate. Every scenario must have one — a missing entry fails
// the test — so the scenario list and this table cannot silently drift apart.
//
// The expectations are stated per signal rather than as one combined verdict,
// which is also the shape of the payload: what each scenario demonstrates is a
// particular COMBINATION of independent signals (env naming a vendor with no
// ancestor to back it, an extension host with no agent process, and so on).
type scenarioExpectation struct {
	envVendors     []string // non-generic env vendors, in any order; nil means none
	ancestorVendor string   // "" means AgentAncestor must be nil
	ideHostVendor  string   // "" means IDEHost must be nil
	stoppedAt      string   // "" means don't check
	wantCI         bool
}

var scenarioExpectations = map[string]scenarioExpectation{
	"direct":            {envVendors: []string{"claude-code"}, ancestorVendor: "claude-code"},
	"inherited-env":     {envVendors: []string{"claude-code"}},
	"ci-image":          {envVendors: []string{"claude-code"}, wantCI: true},
	"node-hosted":       {ancestorVendor: "claude-code"},
	"node-hosted-blind": {},
	"interpreter-gap":   {},
	"pid-reuse":         {stoppedAt: "pid_reuse"},
	"unknown-vendor":    {ancestorVendor: "aider"},
	"wrapper-chain":     {ancestorVendor: "claude-code"},
	"timeout-wrapped":   {envVendors: []string{"claude-code"}, ancestorVendor: "claude-code"},
	"xargs-fanout":      {ancestorVendor: "claude-code"},
	"ide-terminal":      {ideHostVendor: "cursor"},

	// The in-editor surfaces. Phase 1 collapses all of these to the same IDE
	// signal (an editor is in the ancestry) plus whatever env var is set. The
	// agent process in vscode-claude-chat wears the editor helper's basename, so
	// it is kindIDEHost and not argv-eligible — ancestry no longer names it, and
	// the env var is what carries the vendor. That accepted recall loss is what
	// dropping the extension-host/pty-host role split costs.
	"vscode-claude-chat": {
		envVendors: []string{"claude-code"}, ideHostVendor: "vscode",
	},
	"vscode-copilot-chat": {
		envVendors: []string{"github-copilot"}, ideHostVendor: "vscode",
	},
	"cursor-agent-chat": {
		envVendors: []string{"cursor"}, ideHostVendor: "cursor",
	},
	"ide-integrated-terminal": {
		ideHostVendor: "vscode",
	},
	"ide-terminal-stale-env": {
		envVendors: []string{"claude-code"}, ideHostVendor: "vscode",
	},

	"nested-agents": {envVendors: []string{"claude-code"}, ancestorVendor: "codex"},
	"ssh":           {stoppedAt: "remote_boundary"},
	"container":     {envVendors: []string{"claude-code"}},
	"human":         {},
}

func TestScenariosMatchDocumentedOutcome(t *testing.T) {
	for _, s := range scenarios {
		t.Run(s.Name, func(t *testing.T) {
			want, ok := scenarioExpectations[s.Name]
			if !ok {
				t.Fatalf("no expectation defined for scenario %q; add one to scenarioExpectations", s.Name)
			}
			// The scenario's own description is the statement of what this tree is
			// supposed to demonstrate, so a failure reports it rather than leaving
			// the reader to go look the name up.
			t.Logf("scenario %q: %s", s.Name, s.Desc)

			src, start := s.source()
			res := Detect(Options{
				Source: src, StartPid: start, Getenv: s.getenv(),
				IsTerminal: notATTY, KeepChain: true,
			})

			if got := envVendors(res); !sameSet(got, want.envVendors) {
				t.Errorf("env vendors = %v, want %v", got, want.envVendors)
			}

			switch {
			case want.ancestorVendor == "" && res.Signals.AgentAncestor != nil:
				t.Errorf("ancestor = %+v, want nil", res.Signals.AgentAncestor)
			case want.ancestorVendor != "" && res.Signals.AgentAncestor == nil:
				t.Errorf("ancestor = nil, want vendor %q", want.ancestorVendor)
			case want.ancestorVendor != "" && res.Signals.AgentAncestor.Vendor != want.ancestorVendor:
				t.Errorf("ancestor vendor = %q, want %q", res.Signals.AgentAncestor.Vendor, want.ancestorVendor)
			}

			switch {
			case want.ideHostVendor == "" && res.Signals.IDEHost != nil:
				t.Errorf("ide_host = %+v, want nil", res.Signals.IDEHost)
			case want.ideHostVendor != "" && res.Signals.IDEHost == nil:
				t.Errorf("ide_host = nil, want vendor %q", want.ideHostVendor)
			case want.ideHostVendor != "" && res.Signals.IDEHost.Vendor != want.ideHostVendor:
				t.Errorf("ide_host vendor = %q, want %q", res.Signals.IDEHost.Vendor, want.ideHostVendor)
			}

			if want.stoppedAt != "" && res.Walk.StoppedAt != want.stoppedAt {
				t.Errorf("stopped_at = %q, want %q", res.Walk.StoppedAt, want.stoppedAt)
			}

			if want.wantCI && len(res.Signals.CI) == 0 {
				t.Error("ci = [], want at least one CI var reported")
			}
			if !want.wantCI && len(res.Signals.CI) != 0 {
				t.Errorf("ci = %v, want none", res.Signals.CI)
			}
		})
	}

	// Drift check in the other direction: an expectation for a name that no longer
	// exists means the scenario list changed out from under this table.
	for name := range scenarioExpectations {
		if _, ok := findScenario(name); !ok {
			t.Errorf("scenarioExpectations has %q but the scenarios list does not", name)
		}
	}
}

// sameSet compares without ordering, because the order of env matches is an
// artifact of the fingerprint table and nothing should depend on it.
func sameSet(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	seen := make(map[string]int, len(got))
	for _, g := range got {
		seen[g]++
	}
	for _, w := range want {
		if seen[w] == 0 {
			return false
		}
		seen[w]--
	}
	return true
}

// Versioned and bitness-suffixed basenames resolve to their table stem by rule
// rather than by enumeration, so an interpreter version or a JetBrains launcher we
// have never seen is typed correctly on first contact. The name reported is the
// STEM, because that is the table key — the observed basename is not vocabulary.
func TestVersionSuffixResolvesToTableStem(t *testing.T) {
	cases := map[string]struct {
		kind   procKind
		vendor string
		key    string
	}{
		"python3":    {kindInterpreter, "", "python"},
		"python3.13": {kindInterpreter, "", "python"},
		"python3.15": {kindInterpreter, "", "python"}, // a version that does not exist yet
		"idea64":     {kindIDEHost, "jetbrains", "idea"},
		"rider64":    {kindIDEHost, "jetbrains", "rider"}, // never had a row of its own
		"datagrip64": {kindIDEHost, "jetbrains", "datagrip"},
	}
	for name, want := range cases {
		fp, key, ok := lookupFingerprint(name)
		if !ok {
			t.Errorf("lookupFingerprint(%q): no match", name)
			continue
		}
		if fp.Kind != want.kind {
			t.Errorf("lookupFingerprint(%q).Kind = %q, want %q", name, fp.Kind, want.kind)
		}
		if fp.Vendor != want.vendor {
			t.Errorf("lookupFingerprint(%q).Vendor = %q, want %q", name, fp.Vendor, want.vendor)
		}
		if key != want.key {
			t.Errorf("lookupFingerprint(%q) key = %q, want the table stem %q", name, key, want.key)
		}
	}

	// The rule must not manufacture attributions. A version-named binary trims to
	// nothing, and trimming digits off an unrecognized name must never land on an
	// agent row — a fabricated vendor is the one error the precision guards exist
	// to prevent.
	for _, name := range []string{"2.1.219", "1.0", "42", "aider2", "claude3", "q1", "codex64"} {
		if fp, _, ok := lookupFingerprint(name); ok {
			t.Errorf("lookupFingerprint(%q) = %+v, want no match", name, fp)
		}
	}
}

func TestNormalizeName(t *testing.T) {
	cases := map[string]string{
		`C:\Program Files\nodejs\node.exe`: "node",
		"/usr/local/bin/Claude":            "claude",
		"  node  ":                         "node",
		// npm ships the native binary as bin/claude.exe on EVERY platform, so
		// stripping .exe on Unix is load-bearing, not just Windows hygiene.
		"/opt/node_modules/@anthropic-ai/claude-code/bin/claude.exe": "claude",
		// The version-named native install. Nothing to match; documented so the
		// behaviour is deliberate rather than accidental.
		"/Users/x/.local/share/claude/versions/2.1.220": "2.1.220",
	}
	for in, want := range cases {
		if got := normalizeName(in); got != want {
			t.Errorf("normalizeName(%q) = %q, want %q", in, got, want)
		}
	}
}

// Every pattern in cmdlineFingerprints is a package or install path written with
// forward slashes, and on Windows the same install presents with backslashes. So
// the whole multi-segment half of the table — which is most of its recall —
// matched nothing on Windows until matchCmdline normalized separators.
//
// This was found by a Windows IDE fixture, not by design. Worth its own test
// because the bug has nothing to do with editors: it hits any Windows chain,
// including a plain `node ...\claude-code\cli.js` under cmd.exe.
func TestWindowsPathSeparatorsStillMatchPatterns(t *testing.T) {
	for _, tt := range []struct {
		name   string
		argv   []string
		vendor string
	}{
		{
			name:   "claude-code under node",
			argv:   []string{`C:\Program Files\nodejs\node.exe`, `C:\Users\x\AppData\npm\node_modules\@anthropic-ai\claude-code\cli.js`},
			vendor: "claude-code",
		},
		{
			name:   "aider under python",
			argv:   []string{`C:\Python313\python.exe`, `C:\Python313\Lib\site-packages\aider\main.py`},
			vendor: "aider",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// Names are pre-normalized: fakeSource returns ProcInfo verbatim, and
			// normalizeName runs in the ProcSource (see ProcInfo.Name). Only the
			// argv is Windows-shaped, which is the thing under test.
			src := fakeSource{
				1000: {Pid: 1000, Ppid: 1001, Name: "cmd"},
				1001: {Pid: 1001, Ppid: 1, Name: normalizeName(tt.argv[0]), Cmdline: tt.argv},
			}
			res := detect(t, src, 1000, nil)
			a := res.Signals.AgentAncestor
			if a == nil {
				t.Fatalf("no agent ancestor for %v; a backslash path must match the same pattern a slash path does", tt.argv)
			}
			if a.Vendor != tt.vendor {
				t.Errorf("vendor = %q, want %q", a.Vendor, tt.vendor)
			}
		})
	}
}

// Normalization is the walk's job, not the source's, and this is the test that
// keeps it there. Every name below is a shape some platform actually hands back —
// a darwin exec path, a Windows exec path, a capitalized comm, a bare .exe — and
// each one has to reach the same fingerprint row as the clean basename does.
//
// The failure mode is why this is enforced structurally rather than documented.
// A source that skips normalization returns no error and drops no field; it just
// misses every lookup, and a miss is indistinguishable from a machine with no
// agent on it. Two Windows fixtures written against the old comment-only contract
// broke it on first attempt.
func TestWalkNormalizesNamesFromAnySource(t *testing.T) {
	for _, tt := range []struct {
		name   string
		raw    string
		vendor string
	}{
		{name: "unix exec path", raw: "/opt/homebrew/bin/claude", vendor: "claude-code"},
		{name: "windows exec path", raw: `C:\Users\x\AppData\Local\Programs\claude\claude.exe`, vendor: "claude-code"},
		{name: "bare exe", raw: "claude.exe", vendor: "claude-code"},
		{name: "capitalized comm", raw: "Claude", vendor: "claude-code"},
		{name: "surrounding whitespace", raw: "  claude\n", vendor: "claude-code"},
		{name: "already normalized", raw: "claude", vendor: "claude-code"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			src := fakeSource{
				1000: {Pid: 1000, Ppid: 1001, Name: "zsh"},
				1001: {Pid: 1001, Ppid: 1, Name: tt.raw},
			}
			a := detect(t, src, 1000, nil).Signals.AgentAncestor
			if a == nil {
				t.Fatalf("no agent ancestor for source name %q; the walk must normalize whatever a source hands it", tt.raw)
			}
			if a.Vendor != tt.vendor {
				t.Errorf("vendor = %q, want %q", a.Vendor, tt.vendor)
			}
			// And the emitted name is still the table KEY, never the raw
			// observation — normalizing a path must not become a way to launder one
			// into the payload.
			if a.Name != "claude" {
				t.Errorf("name = %q, want the table key %q", a.Name, "claude")
			}
		})
	}
}

// The platform-shape rules have to survive a raw path too: a versioned
// interpreter and an Electron helper both arrive as full exec paths in practice,
// and both resolve through rules rather than exact rows.
func TestWalkNormalizesNamesBeforeTheShapeRules(t *testing.T) {
	versioned := fakeSource{
		1000: {Pid: 1000, Ppid: 1001, Name: "bash"},
		1001: {Pid: 1001, Ppid: 1, Name: "/usr/bin/python3.13", Cmdline: []string{
			"/usr/bin/python3.13", "/usr/lib/python3.13/site-packages/aider/main.py",
		}},
	}
	if a := detect(t, versioned, 1000, nil).Signals.AgentAncestor; a == nil || a.Vendor != "aider" {
		t.Errorf("ancestor = %+v, want aider — python3.13 must reach the python row from a full path", a)
	}

	helper := fakeSource{
		1000: {Pid: 1000, Ppid: 1001, Name: "zsh"},
		1001: {Pid: 1001, Ppid: 1002, Name: "/Applications/Visual Studio Code.app/Contents/Frameworks/Code Helper (Plugin).app/Contents/MacOS/Code Helper (Plugin)"},
		1002: {Pid: 1002, Ppid: 1, Name: "/Applications/Visual Studio Code.app/Contents/MacOS/Electron"},
	}
	if h := detect(t, helper, 1000, nil).Signals.IDEHost; h == nil || h.Vendor != "vscode" {
		t.Errorf("ide_host = %+v, want vscode from a full helper path", h)
	}
}

// Every non-generic env-marker key should have an entry in envVendorForKey, so
// the test suite's vendor-named assertions stay meaningful and don't silently
// stop covering a row added to envFingerprints.
func TestEnvVendorForKeyCoversAllMarkers(t *testing.T) {
	for _, v := range envFingerprints {
		if v == "AI_AGENT" || v == "AGENT" {
			continue
		}
		if _, ok := envVendorForKey[v]; !ok {
			t.Errorf("%s: no entry in envVendorForKey", v)
		}
	}
}
