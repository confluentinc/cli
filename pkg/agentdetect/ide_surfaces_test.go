package agentdetect

import "testing"

// IDE surfaces. macOS chains are captured from real process trees; the Linux
// chain is MODELED from VS Code's process model, not captured yet.
//
// One IDE signal: an editor anywhere in the ancestry means "agent-capable
// environment" (Signals.IDEHost); an agent wearing the helper's name is caught by
// its env var, not ancestry (see TestInEditorAgentUnderHelperIsNotAttributed).

// A VS Code helper in the ancestry is an IDE host even though `code` itself is
// deeper up. The helper resolves to its editor's key, so the recorded name is
// "code", not the observed helper basename.
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
	// Nearest editor wins, recorded as the editor's key — which also reaches the wire.
	if res.Signals.IDEHost.Depth != 2 {
		t.Errorf("ide_host depth = %d, want 2", res.Signals.IDEHost.Depth)
	}
	if res.Signals.IDEHost.Name != "code" {
		t.Errorf("ide_host name = %q, want %q", res.Signals.IDEHost.Name, "code")
	}
	if got := res.Attributes().IDEHost; got == nil || *got != "code" {
		t.Errorf("attributes ide_host = %v, want %q", got, "code")
	}
}

// The recall cost of dropping in-editor role detection, pinned as a deliberate
// choice: an agent running under the editor helper's basename is kindIDEHost, not
// argv-eligible, so ancestry can't name it — the env var it sets is what catches it.
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

// A known editor's helper matches regardless of role (any "(role)" is discarded);
// Cursor/Windsurf forks are covered by their own rows. Gated on a known editor —
// see TestNonEditorElectronHelpersAreNotIDEHosts.
func TestHelperClassificationMatchesKnownEditors(t *testing.T) {
	editors := map[string]string{ // helper basename → expected vendor
		"code helper (plugin)":     "vscode",
		"cursor helper (plugin)":   "cursor",
		"windsurf helper (plugin)": "windsurf",
		"code helper":              "vscode",
		"code helper (renderer)":   "vscode",
	}
	for name, wantVendor := range editors {
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
}

// " helper" is an Electron convention, not an editor one, so non-editor Electron
// apps (Slack, Discord, the Hyper terminal) and unknown editor forks must not
// become IDE hosts. Bare "helper" is not a match either.
func TestNonEditorElectronHelpersAreNotIDEHosts(t *testing.T) {
	for _, name := range []string{
		"slack helper (renderer)",
		"discord helper (gpu)",
		"hyper helper",
		"someeditor helper (plugin)", // an editor, but one we have no row for
		"helper",
		"helper (plugin)",
		"myhelper",
	} {
		if fp, _, ok := lookupFingerprint(name); ok && fp.Kind == kindIDEHost {
			t.Errorf("lookupFingerprint(%q) = %+v, must not classify as an IDE host", name, fp)
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

// On Linux/Windows the editor re-execs one `code` binary per child role, so an
// in-editor agent's parent is basename `code` → kindIDEHost, not argv-eligible.
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

// Argv matching covers kindUnknown ancestors & interpreters. For an
// unknown basename that means argv[0] only — its other args are user data, so
// `rg @anthropic-ai/claude-code` isn't counted as an agent.
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

	// The given-up half: the same path as an ARGUMENT of an unknown binary is not attributed.
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

	// Unmatched: counted as an unknown gap, kept separate from the interpreter gap.
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

// Identity resolution beyond the executable basename. Regression tests for a real
// miss: Claude Code's native install left agent_ancestor null. Observed values.

// The install that motivated the rule; argv[0] resolves it here, the cheap path.
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

// When argv is unreadable, the path rule has to carry it. Same install, no argv.
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

// The trigger is "basename is a version", not "basename unrecognized" — otherwise a
// username or checkout named like an agent would be attributed.
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

// A path segment is weak evidence; typing a kindIDEHost/kindWrapper process would
// strip its argv eligibility. Restricting to argv-eligible kinds means the rule
// can only ever add an attribution.
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

// argv[0] is a weak identity claim; it must not outrank a resolved basename.
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

// A login shell's argv[0] ("-zsh") must resolve to nothing.
func TestLoginShellArgv0IsHarmless(t *testing.T) {
	if _, _, ok := lookupFingerprint(normalizeName("-zsh")); ok {
		t.Error(`normalizeName("-zsh") resolved; it should match nothing`)
	}
}

// Two hops reaches <product>/versions/<ver>; three would read unrelated dirs.
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
