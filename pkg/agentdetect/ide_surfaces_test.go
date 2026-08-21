package agentdetect

import "testing"

// IDE surfaces. The macOS chains are transcribed from real runs on
// darwin/arm64, 2026-07-31; the Linux chain is MODELED from VS Code's
// documented process model (Chromium re-execs one binary per child role), not
// observed.
//
// Phase 1 keeps a single IDE signal: an editor anywhere in the ancestry means
// "agent-capable environment" (Signals.IDEHost). We deliberately do NOT try to
// tell an agent tool call apart from a human in the integrated terminal — that
// distinction only exists on macOS, needs VS-Code-fork-specific process-role
// detection, and is ambiguous even when it fires. The recall cost is pinned as
// a test below (TestInEditorAgentUnderHelperIsNotAttributed): an agent wearing
// the editor helper's name is caught by its env var, not by ancestry.

// A VS Code editor helper in the ancestry is reported as an IDE host, even
// though the editor's own `code` process is deeper up. Classification is by the
// Electron " helper" suffix, so a fork we have never seen is still recognized.
func TestVSCodeHelperInAncestryIsIDEHost(t *testing.T) {
	src := fakeSource{
		76073: {Pid: 76073, Ppid: 76067, Name: "agentdetect"},
		76067: {Pid: 76067, Ppid: 75809, Name: "zsh"},
		75809: {Pid: 75809, Ppid: 64305, Name: "code helper (plugin)"},
		64305: {Pid: 64305, Ppid: 1, Name: "code"},
	}
	res := detect(t, src, 76067, nil)

	if res.Signals.IDEHost == nil || res.Signals.IDEHost.Vendor != "vscode" {
		t.Fatalf("ide_host = %+v, want vscode", res.Signals.IDEHost)
	}
	// The helper is the nearest editor process, so it is what IDEHost records.
	if res.Signals.IDEHost.Depth != 2 {
		t.Errorf("ide_host depth = %d, want 2", res.Signals.IDEHost.Depth)
	}
}

// The recall cost of dropping in-editor role detection, pinned so it stays a
// deliberate choice rather than a silent regression. An agent whose process
// wears the editor helper's basename — VS Code runs extensions and the node
// children they spawn as the Electron helper — is kindIDEHost, NOT
// argv-eligible, so its argv is never read and ancestry cannot name it. The
// env var it sets is the signal that catches it.
func TestInEditorAgentUnderHelperIsNotAttributed(t *testing.T) {
	src := fakeSource{
		76073: {Pid: 76073, Ppid: 76067, Name: "agentdetect"},
		76067: {Pid: 76067, Ppid: 75809, Name: "zsh"},
		75809: {Pid: 75809, Ppid: 64305, Name: "code helper (plugin)", Cmdline: []string{
			"/Applications/Visual Studio Code.app/Contents/Frameworks/Code Helper (Plugin).app/Contents/MacOS/Code Helper (Plugin)",
			"/Users/x/.vscode/extensions/anthropic.claude-code-2.1.220-darwin-arm64/resources/claude-code/cli.js",
		}},
		64305: {Pid: 64305, Ppid: 1, Name: "code"},
	}
	res := detect(t, src, 76067, map[string]string{"CLAUDECODE": "1"})

	if res.Signals.AgentAncestor != nil {
		t.Errorf("agent_ancestor = %+v, want nil — a helper is not argv-eligible", res.Signals.AgentAncestor)
	}
	if !hasEnvVendor(res, "claude-code") {
		t.Errorf("env vendors = %v, want claude-code (the signal that covers this)", envVendors(res))
	}
	if res.Signals.IDEHost == nil || res.Signals.IDEHost.Vendor != "vscode" {
		t.Errorf("ide_host = %+v, want vscode", res.Signals.IDEHost)
	}
}

// Helper classification is by the Electron " helper" suffix, not an enumerated
// list, so a VS Code fork we have never seen is still an IDE host on first
// contact — with its vendor where an editor row supplies one, empty otherwise.
// Any "(role)" suffix is discarded: Phase 1 keeps only "in an editor".
func TestHelperClassificationGeneralizesAcrossForks(t *testing.T) {
	forks := map[string]string{ // helper basename → expected vendor
		"code helper (plugin)":       "vscode",
		"cursor helper (plugin)":     "cursor",
		"windsurf helper (plugin)":   "windsurf",
		"code helper":                "vscode",
		"code helper (renderer)":     "vscode",
		"someeditor helper (plugin)": "", // unknown fork: known IDE host, unknown vendor
	}
	for name, wantVendor := range forks {
		fp, _, ok := lookupFingerprint(name)
		if !ok {
			t.Errorf("lookupFingerprint(%q): no match", name)
			continue
		}
		if fp.Kind != kindIDEHost {
			t.Errorf("lookupFingerprint(%q).Kind = %q, want %q", name, fp.Kind, kindIDEHost)
		}
		if fp.Vendor != wantVendor {
			t.Errorf("lookupFingerprint(%q).Vendor = %q, want %q", name, fp.Vendor, wantVendor)
		}
	}

	// Must not swallow unrelated names that merely contain the word.
	for _, name := range []string{"helper", "helper (plugin)", "myhelper"} {
		if fp, _, ok := lookupFingerprint(name); ok && fp.Kind == kindIDEHost {
			t.Errorf("lookupFingerprint(%q) = %+v, must not classify as an editor helper", name, fp)
		}
	}
}

// The editor's own process is an IDE host, and the chain shape marks it 'e'.
func TestEditorMainProcessIsIDEHost(t *testing.T) {
	src := fakeSource{
		1000: {Pid: 1000, Ppid: 1001, Name: "zsh"},
		1001: {Pid: 1001, Ppid: 1, Name: "code", Cmdline: []string{"/usr/share/code/code", "."}},
	}
	res := detect(t, src, 1000, nil)

	if res.Signals.IDEHost == nil || res.Signals.IDEHost.Vendor != "vscode" {
		t.Errorf("ide_host = %+v, want vscode", res.Signals.IDEHost)
	}
	if got := res.Signals.ChainShape; got != "se" {
		t.Errorf("chain shape = %q, want %q", got, "se")
	}
}

// On Linux/Windows the editor re-execs one `code` binary for every child role,
// so an in-editor agent's parent is basename `code` → kindIDEHost, not
// argv-eligible. Phase 1 makes no attempt to read the role out of --type; the
// behavior matches macOS and the env var carries the vendor.
func TestLinuxInEditorIsIDEHostAndEnvCarriesVendor(t *testing.T) {
	src := fakeSource{
		1000: {Pid: 1000, Ppid: 1001, Name: "agentdetect"},
		1001: {Pid: 1001, Ppid: 1002, Name: "bash"},
		1002: {Pid: 1002, Ppid: 1003, Name: "code", Cmdline: []string{
			"/usr/share/code/code",
			"/home/x/.vscode/extensions/anthropic.claude-code/resources/claude-code/cli.js",
			"--type=utility", "--utility-sub-type=node.mojom.NodeService",
		}},
		1003: {Pid: 1003, Ppid: 1, Name: "code"},
	}
	res := detect(t, src, 1001, map[string]string{"CLAUDECODE": "1"})

	if res.Signals.AgentAncestor != nil {
		t.Errorf("agent_ancestor = %+v, want nil — `code` is not argv-eligible", res.Signals.AgentAncestor)
	}
	if !hasEnvVendor(res, "claude-code") {
		t.Errorf("env vendors = %v, want claude-code", envVendors(res))
	}
	if res.Signals.IDEHost == nil || res.Signals.IDEHost.Vendor != "vscode" {
		t.Errorf("ide_host = %+v, want vscode", res.Signals.IDEHost)
	}
}

// Argv matching covers unknown-kind ancestors, not just interpreters — otherwise
// any agent shipped under a name we have no fingerprint for is invisible even
// when its argv names it outright. A miss is tagged with its kind so the two
// populations stay distinguishable in telemetry.
//
// For an unknown basename that means argv[0] only: an unrecognized binary is far
// more often an ordinary tool operating ON a path than an agent, so its arguments
// are user data. The recall this gives up — an unrecognized LAUNCHER whose child
// agent is named in an argument — is real, and it is the trade that keeps
// `rg @anthropic-ai/claude-code` from being counted as an agent call. What is lost
// still shows up in the unattributed-unknown count rather than vanishing, so the
// size of the trade is measurable from production instead of assumed.
func TestUnknownAncestorsAreArgvMatchedAndCounted(t *testing.T) {
	// Matched: an unrecognized binary invoked by its own agent install path.
	src := fakeSource{
		1000: {Pid: 1000, Ppid: 1001, Name: "zsh"},
		1001: {Pid: 1001, Ppid: 1, Name: "some-unknown-launcher", Cmdline: []string{
			"/opt/node_modules/@openai/codex/bin/some-unknown-launcher", "--headless",
		}},
	}
	res := detect(t, src, 1000, nil)
	if res.Signals.AgentAncestor == nil || res.Signals.AgentAncestor.Vendor != "codex" {
		t.Fatalf("ancestor = %+v, want codex via argv[0]", res.Signals.AgentAncestor)
	}

	// The given-up half, asserted so the trade is deliberate rather than accidental:
	// the same agent path in an ARGUMENT of an unknown binary is not attributed.
	inArg := fakeSource{
		1000: {Pid: 1000, Ppid: 1001, Name: "zsh"},
		1001: {Pid: 1001, Ppid: 1, Name: "some-unknown-launcher", Cmdline: []string{
			"some-unknown-launcher", "/opt/node_modules/@openai/codex/bin/codex.js",
		}},
	}
	res = detect(t, inArg, 1000, nil)
	if res.Signals.AgentAncestor != nil {
		t.Errorf("ancestor = %+v, want nil; an unknown binary's arguments are user data",
			res.Signals.AgentAncestor)
	}
	if len(res.Signals.Unattributed) != 1 || res.Signals.Unattributed[0].Kind != kindUnknown {
		t.Errorf("unattributed = %+v, want one unknown — what precision costs must stay countable",
			res.Signals.Unattributed)
	}

	// Unmatched: counted as an unknown gap, kept apart from the interpreter gap
	// so "our table is missing a vendor" and "we saw a process we can't type at
	// all" remain separate numbers.
	miss := fakeSource{
		1000: {Pid: 1000, Ppid: 1001, Name: "zsh"},
		1001: {Pid: 1001, Ppid: 1, Name: "some-unknown-launcher", Cmdline: []string{
			"some-unknown-launcher", "--do-a-thing",
		}},
	}
	res = detect(t, miss, 1000, nil)
	if res.Signals.AgentAncestor != nil {
		t.Errorf("ancestor = %+v, want nil", res.Signals.AgentAncestor)
	}
	if len(res.Signals.Unattributed) != 1 {
		t.Fatalf("unattributed = %+v, want exactly one", res.Signals.Unattributed)
	}
	u := res.Signals.Unattributed[0]
	if u.Kind != kindUnknown {
		t.Errorf("kind = %q, want %q (not an interpreter — the distinction routes the fix)",
			u.Kind, kindUnknown)
	}
	if u.Depth != 2 {
		t.Errorf("depth = %d, want 2", u.Depth)
	}
	if !u.ArgvReadable {
		t.Error("argv_readable = false, want true")
	}
}

// Identity resolution beyond the executable basename.
//
// These are regression tests for a miss found by the field probe, not a
// hypothetical: Claude Code's native install put the agent three frames up the
// real ancestry with agent_ancestor null. The exact observed values are used.

// The install that motivated the whole rule. argv[0] resolves it, which is the
// cheap path; TestVersionedPathResolvesWhenArgvIsUnreadable covers the case
// where it does not.
func TestNativeInstallWithAVersionedBasenameIsAttributed(t *testing.T) {
	src := fakeSource{
		1000: {Pid: 1000, Ppid: 1001, Name: "/bin/zsh", Cmdline: []string{"/bin/zsh"}},
		1001: {
			Pid: 1001, Ppid: 1,
			Name:    "/Users/someone/.local/share/claude/versions/2.1.231",
			Cmdline: []string{"claude"},
		},
	}
	res := detect(t, src, 1000, nil)

	a := res.Signals.AgentAncestor
	if a == nil {
		t.Fatalf("agent_ancestor = nil, want claude-code; chain_shape = %q", res.Signals.ChainShape)
	}
	if a.Vendor != "claude-code" || a.Name != "claude" {
		t.Errorf("agent = %+v, want vendor claude-code and key claude", a)
	}
	if got := res.Signals.ChainShape; got != "sa" {
		t.Errorf("chain_shape = %q, want %q", got, "sa")
	}
}

// argv is permission-gated, and a rewritten or unreadable argv is exactly when
// the path rule has to carry it. Same install, no argv.
func TestVersionedPathResolvesWhenArgvIsUnreadable(t *testing.T) {
	src := fakeSource{
		1000: {Pid: 1000, Ppid: 1001, Name: "/bin/zsh"},
		1001: {Pid: 1001, Ppid: 1, Name: "/Users/someone/.local/share/claude/versions/2.1.231"},
	}
	res := detect(t, src, 1000, nil)

	a := res.Signals.AgentAncestor
	if a == nil || a.Vendor != "claude-code" {
		t.Fatalf("agent_ancestor = %+v, want claude-code from the path", a)
	}

	// Windows installs the same shape with the other separator and a suffix.
	src = fakeSource{
		1000: {Pid: 1000, Ppid: 1001, Name: `C:\Windows\System32\cmd.exe`},
		1001: {Pid: 1001, Ppid: 1, Name: `C:\Users\Someone\AppData\Local\claude\versions\2.1.231.exe`},
	}
	if a := detect(t, src, 1000, nil).Signals.AgentAncestor; a == nil || a.Vendor != "claude-code" {
		t.Errorf("windows: agent_ancestor = %+v, want claude-code", a)
	}
}

// The trigger is "the basename is a version", not "the basename is unrecognized".
// A rule without that restriction attributes an agent to anyone whose username or
// checkout happens to be named like one — straight into the headline number.
func TestPathSegmentsAreOnlyReadForAVersionedBasename(t *testing.T) {
	for _, path := range []string{
		"/Users/claude/bin/mytool",       // username
		"/home/aider/scripts/run.sh",     // username
		"/Users/x/dev/cursor/build/tool", // checkout
		"/opt/claude/bin/helper",         // sibling binary in an agent's own dir
	} {
		src := fakeSource{
			1000: {Pid: 1000, Ppid: 1001, Name: "/bin/bash"},
			1001: {Pid: 1001, Ppid: 1, Name: path},
		}
		res := detect(t, src, 1000, nil)
		if a := res.Signals.AgentAncestor; a != nil {
			t.Errorf("%s: agent_ancestor = %+v, want none — the basename is not a version", path, a)
		}
	}
}

// A path segment is weaker evidence than a basename, and a wrong one that types a
// process kindIDEHost or kindWrapper would REMOVE its argv eligibility, turning a
// counted unknown into an invisible miss. Restricting the rule to argv-eligible
// kinds means it can only ever add an attribution.
func TestPathSegmentsCannotDeleteASignal(t *testing.T) {
	// "code" is a real table key, and an ineligible one.
	src := fakeSource{
		1000: {Pid: 1000, Ppid: 1001, Name: "/bin/bash"},
		1001: {Pid: 1001, Ppid: 1, Name: "/opt/code/versions/1.2.3"},
	}
	res := detect(t, src, 1000, nil)

	if res.Signals.IDEHost != nil {
		t.Errorf("ide_host = %+v, want none from a path segment", res.Signals.IDEHost)
	}
	if got := res.Signals.ChainShape; got != "s?" {
		t.Errorf("chain_shape = %q, want %q — it must stay unknown and eligible", got, "s?")
	}
	if len(res.Signals.Unattributed) != 1 {
		t.Errorf("unattributed = %+v, want the process still counted", res.Signals.Unattributed)
	}
}

// argv[0] is an identity claim, but a weak one, and it must not outrank a real
// basename that already resolved.
func TestArgv0DoesNotOverrideAResolvedBasename(t *testing.T) {
	src := fakeSource{
		1000: {Pid: 1000, Ppid: 1001, Name: "/bin/zsh", Cmdline: []string{"claude"}},
		1001: {Pid: 1001, Ppid: 1, Name: "/sbin/launchd"},
	}
	res := detect(t, src, 1000, nil)

	if a := res.Signals.AgentAncestor; a != nil {
		t.Errorf("agent_ancestor = %+v, want none: the basename says zsh and argv is its command", a)
	}
	if got := res.Signals.ChainShape; got != "sn" {
		t.Errorf("chain_shape = %q, want %q", got, "sn")
	}
}

// A login shell's argv[0] is "-zsh". It must not resolve to anything, and must
// not stop the real basename from having already won.
func TestLoginShellArgv0IsHarmless(t *testing.T) {
	if _, _, ok := lookupFingerprint(normalizeName("-zsh")); ok {
		t.Error(`normalizeName("-zsh") resolved; it should match nothing`)
	}
}

// Two hops reaches <product>/versions/<version> and stops there. Three would
// start reading directories that have nothing to do with the program.
func TestPathSegmentSearchIsBounded(t *testing.T) {
	src := fakeSource{
		1000: {Pid: 1000, Ppid: 1001, Name: "/bin/bash"},
		1001: {Pid: 1001, Ppid: 1, Name: "/opt/claude/releases/stable/x64/1.2.3"},
	}
	if a := detect(t, src, 1000, nil).Signals.AgentAncestor; a != nil {
		t.Errorf("agent_ancestor = %+v, want none — claude is %d hops up, past the bound",
			a, 4)
	}
}
