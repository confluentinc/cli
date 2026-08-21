package agentdetect

import "testing"

// Chains transcribed from real runs on darwin/arm64, 2026-07-31. Unlike the
// synthetic trees in detect_test.go these are not illustrative — each one is a
// process tree that was observed, with the pids and argv it actually had. They
// exist because the first three IDE-chat surfaces we tried all reported
// "no agent ancestor matched" and landed in the env_only bucket, i.e. the
// bucket the memo cites as the env-var false-positive case.
//
// The shared mechanism: VS Code and its forks run the extension host — and any
// node process an extension spawns — as the Electron helper binary, so the
// executable basename is "code helper (plugin)" / "cursor helper (plugin)",
// never "node". Nothing in the tier-1 table matched, the ancestor was typed
// unknown, and because tier-2 matching was gated on kindInterpreter the argv
// was never consulted even where it named the vendor outright.

// Surface 1: Claude Code's VS Code extension. The extension launches the real
// CLI as a child of the extension host (extension.js falls back to
// process.execPath + bundled resources/claude-code/cli.js), so there IS a
// distinct agent process — it is just wearing the Electron helper's name. Its
// argv names the vendor, so tier 2 can recover it once the kind gate allows it.
func TestVSCodeClaudeChatIsAttributedFromExtHostArgv(t *testing.T) {
	src := fakeSource{
		76073: {Pid: 76073, Ppid: 76067, Name: "agentdetect"},
		76067: {Pid: 76067, Ppid: 75809, Name: "zsh"},
		75809: {Pid: 75809, Ppid: 65080, Name: "code helper (plugin)", Cmdline: []string{
			"/Applications/Visual Studio Code.app/Contents/Frameworks/Code Helper (Plugin).app/Contents/MacOS/Code Helper (Plugin)",
			"/Users/x/.vscode/extensions/anthropic.claude-code-2.1.220-darwin-arm64/resources/claude-code/cli.js",
		}},
		65080: {Pid: 65080, Ppid: 64305, Name: "code helper (plugin)", Cmdline: []string{
			"/Applications/Visual Studio Code.app/Contents/Frameworks/Code Helper (Plugin).app/Contents/MacOS/Code Helper (Plugin)",
			"--type=utility", "--utility-sub-type=node.mojom.NodeService",
		}},
		64305: {Pid: 64305, Ppid: 1, Name: "code"},
	}
	res := detect(t, src, 76067, map[string]string{"CLAUDECODE": "1"})

	a := res.Signals.AgentAncestor
	if a == nil {
		t.Fatal("no agent ancestor; the extension host's argv names claude-code and must be consulted")
	}
	if a.Vendor != "claude-code" {
		t.Errorf("vendor = %q, want claude-code", a.Vendor)
	}
	if got := a.MatchedOn(); got != "argv" {
		t.Errorf("matched_on = %q, want argv", got)
	}
	if a.Depth != 2 {
		t.Errorf("depth = %d, want 2", a.Depth)
	}
	if !hasEnvVendor(res, "claude-code") {
		t.Errorf("env vendors = %v, want claude-code", envVendors(res))
	}
	// Provenance is recorded independently of vendor attribution: this shell was
	// spawned by an extension host, not by the integrated terminal.
	s := res.Signals.IDESpawn
	if s == nil || s.Via != spawnExtensionHost {
		t.Fatalf("ide_spawn = %+v, want via %q", s, spawnExtensionHost)
	}
	if s.Vendor != "vscode" {
		t.Errorf("ide_spawn vendor = %q, want vscode", s.Vendor)
	}
}

// Surface 2: GitHub Copilot's chat in VS Code. Here the shell's direct parent
// IS the extension host (pid 65080, the same process that appears at depth 3
// above), and its argv is Chromium boilerplate with no vendor anywhere in it.
// There is no distinct agent process at all — so no fingerprint table can name
// the vendor from ancestry. Provenance is still recoverable, and the env var
// remains the only vendor signal.
func TestVSCodeCopilotChatHasNoAgentProcessButKnownProvenance(t *testing.T) {
	src := fakeSource{
		77884: {Pid: 77884, Ppid: 77880, Name: "agentdetect"},
		77880: {Pid: 77880, Ppid: 65080, Name: "bash"},
		65080: {Pid: 65080, Ppid: 64305, Name: "code helper (plugin)", Cmdline: []string{
			"/Applications/Visual Studio Code.app/Contents/Frameworks/Code Helper (Plugin).app/Contents/MacOS/Code Helper (Plugin)",
			"--type=utility", "--utility-sub-type=node.mojom.NodeService", "--lang=en-US",
		}},
		64305: {Pid: 64305, Ppid: 1, Name: "code"},
	}
	res := detect(t, src, 77880, map[string]string{"COPILOT_CLI": "1"})

	if res.Signals.AgentAncestor != nil {
		t.Errorf("ancestor = %+v, want nil — the agent has no process of its own here",
			res.Signals.AgentAncestor)
	}
	s := res.Signals.IDESpawn
	if s == nil || s.Via != spawnExtensionHost || s.Depth != 2 {
		t.Fatalf("ide_spawn = %+v, want via %q at depth 2", s, spawnExtensionHost)
	}
	// The taxonomy claim, and the reason provenance is a separate field: the env
	// var is the ONLY vendor signal here, but it is not bare. An extension host
	// spawned the shell, so the "human who inherited a stale variable" reading is
	// ruled out by the ancestry rather than merely unsupported by it — and the two
	// facts have to be separable downstream for that distinction to be usable.
	if !hasEnvVendor(res, "github-copilot") {
		t.Errorf("env vendors = %v, want github-copilot", envVendors(res))
	}
	// Counted, and tagged with the kind that says which population it belongs to:
	// argv WAS readable and still named no vendor, so this is the structural-miss
	// population that no table update can reach — not a table gap and not an
	// uncharacterized process shape.
	if len(res.Signals.Unattributed) != 1 {
		t.Fatalf("unattributed = %+v, want exactly one", res.Signals.Unattributed)
	}
	u := res.Signals.Unattributed[0]
	if u.Kind != kindIDEExtHost {
		t.Errorf("kind = %q, want %q", u.Kind, kindIDEExtHost)
	}
	if !u.ArgvReadable {
		t.Error("argv_readable = false, want true — argv was read, it just names no vendor")
	}
}

// Surface 3: Cursor's agent. Same shape as Copilot — the agent lives in the
// extension host — with two extra wrinkles. Cursor rewrites the extension
// host's argv to a process title, destroying anything the pattern table could
// have matched; and it interposes `cursorsandbox`, which it applies to agent-run
// commands but not to the integrated terminal.
func TestCursorAgentChatIsSandboxedAndArgvRewritten(t *testing.T) {
	src := fakeSource{
		56331: {Pid: 56331, Ppid: 56330, Name: "agentdetect"},
		56330: {Pid: 56330, Ppid: 56262, Name: "zsh"},
		56262: {Pid: 56262, Ppid: 56260, Name: "zsh"},
		56260: {Pid: 56260, Ppid: 55852, Name: "cursorsandbox"},
		// argv is the rewritten title, not the original command line.
		55852: {Pid: 55852, Ppid: 55169, Name: "cursor helper (plugin)", Cmdline: []string{
			"Cursor Helper (Plugin): extension-host agentdetect-poc [1-2]",
		}},
		55169: {Pid: 55169, Ppid: 1, Name: "cursor"},
	}
	res := detect(t, src, 56330, map[string]string{"CURSOR_AGENT": "1"})

	if res.Signals.AgentAncestor != nil {
		t.Errorf("ancestor = %+v, want nil — argv was rewritten, no vendor left to match",
			res.Signals.AgentAncestor)
	}
	s := res.Signals.IDESpawn
	if s == nil || s.Via != spawnExtensionHost || s.Depth != 4 {
		t.Fatalf("ide_spawn = %+v, want via %q at depth 4", s, spawnExtensionHost)
	}
	if s.Vendor != "cursor" {
		t.Errorf("ide_spawn vendor = %q, want cursor", s.Vendor)
	}
	if !hasEnvVendor(res, "cursor") {
		t.Errorf("env vendors = %v, want cursor", envVendors(res))
	}
	// The sandbox is corroborating evidence, so it has to survive as a wrapper
	// rather than dropping through as an unknown.
	if len(res.Signals.Wrappers) != 1 || res.Signals.Wrappers[0].Name != "cursorsandbox" {
		t.Errorf("wrappers = %+v, want cursorsandbox", res.Signals.Wrappers)
	}
	if res.Signals.Wrappers[0].Depth != 3 {
		t.Errorf("cursorsandbox depth = %d, want 3", res.Signals.Wrappers[0].Depth)
	}
}

// The negative control, and the reason the provenance signal is worth anything:
// a human typing in VS Code's integrated terminal. Observed on the same machine
// in the same session as surfaces 1 and 2. The editor's pty host is a
// NON-plugin helper, so this chain is distinguishable from all three agent
// surfaces above by executable name alone — no argv read required.
func TestHumanIntegratedTerminalIsDistinguishableFromExtHost(t *testing.T) {
	src := fakeSource{
		49595: {Pid: 49595, Ppid: 64563, Name: "zsh"},
		64563: {Pid: 64563, Ppid: 64305, Name: "code helper", Cmdline: []string{
			"/Applications/Visual Studio Code.app/Contents/Frameworks/Code Helper.app/Contents/MacOS/Code Helper",
			"--type=utility", "--utility-sub-type=node.mojom.NodeService",
		}},
		64305: {Pid: 64305, Ppid: 1, Name: "code"},
	}
	res := detect(t, src, 49595, nil)

	if res.Signals.AgentAncestor != nil {
		t.Errorf("ancestor = %+v, want nil", res.Signals.AgentAncestor)
	}
	s := res.Signals.IDESpawn
	if s == nil || s.Via != spawnIDEUtility {
		t.Fatalf("ide_spawn = %+v, want via %q", s, spawnIDEUtility)
	}
	if got := envVendors(res); len(got) != 0 {
		t.Errorf("env vendors = %v, want none", got)
	}

	// And with a stale variable inherited into that same terminal, the env signal
	// fires alone against a NON-extension editor helper — which is what makes the
	// inheritance reading a positive finding rather than a shrug: the ancestry
	// actively supports it instead of merely failing to contradict it.
	stale := detect(t, src, 49595, map[string]string{"CLAUDECODE": "1"})
	if !hasEnvVendor(stale, "claude-code") {
		t.Errorf("env vendors = %v, want claude-code", envVendors(stale))
	}
	if stale.Signals.AgentAncestor != nil {
		t.Errorf("ancestor = %+v, want nil", stale.Signals.AgentAncestor)
	}
	if s := stale.Signals.IDESpawn; s == nil || s.Via != spawnIDEUtility {
		t.Errorf("ide_spawn = %+v, want via %q", s, spawnIDEUtility)
	}
}

// Provenance classification is by suffix, not by an enumerated list of helper
// basenames, so a VS Code fork we have never seen is classified correctly on
// first contact. Enumerating would need a new table entry per editor per
// platform forever; the suffix is a property of how Electron apps are packaged.
func TestHelperClassificationGeneralizesAcrossForks(t *testing.T) {
	cases := map[string]struct {
		kind   procKind
		vendor string
	}{
		"code helper (plugin)":     {kindIDEExtHost, "vscode"},
		"cursor helper (plugin)":   {kindIDEExtHost, "cursor"},
		"windsurf helper (plugin)": {kindIDEExtHost, "windsurf"},
		"code helper":              {kindIDEUtility, "vscode"},
		"cursor helper":            {kindIDEUtility, "cursor"},
		"code helper (renderer)":   {kindIDEUtility, "vscode"},
		"code helper (gpu)":        {kindIDEUtility, "vscode"},
		// A fork whose editor basename we have no fingerprint for: the ROLE is
		// still known even though the vendor is not. Role is what the taxonomy
		// needs, so this is a partial hit, not a miss.
		"someeditor helper (plugin)": {kindIDEExtHost, ""},
	}
	for name, want := range cases {
		fp, _, ok := lookupFingerprint(name)
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
	}

	// Must not swallow unrelated names that merely contain the word.
	for _, name := range []string{"helper", "helper (plugin)", "myhelper", "code"} {
		if fp, _, ok := lookupFingerprint(name); ok && (fp.Kind == kindIDEExtHost || fp.Kind == kindIDEUtility) {
			t.Errorf("lookupFingerprint(%q) = %+v, must not classify as an editor helper", name, fp)
		}
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

// Chain shape has to carry the new kinds, since it is the compact field that
// rides on every invocation and the whole point is that these chains are now
// distinguishable from each other at a glance.
func TestChainShapeCarriesEditorHelperRoles(t *testing.T) {
	agent := fakeSource{
		1000: {Pid: 1000, Ppid: 1001, Name: "bash"},
		1001: {Pid: 1001, Ppid: 1002, Name: "code helper (plugin)"},
		1002: {Pid: 1002, Ppid: 1, Name: "code"},
	}
	if got := detect(t, agent, 1000, nil).Signals.ChainShape; got != "sxe" {
		t.Errorf("ext-host chain shape = %q, want %q", got, "sxe")
	}

	human := fakeSource{
		1000: {Pid: 1000, Ppid: 1001, Name: "zsh"},
		1001: {Pid: 1001, Ppid: 1002, Name: "code helper"},
		1002: {Pid: 1002, Ppid: 1, Name: "code"},
	}
	if got := detect(t, human, 1000, nil).Signals.ChainShape; got != "sue" {
		t.Errorf("integrated-terminal chain shape = %q, want %q", got, "sue")
	}
}

// ---------------------------------------------------------------------------
// Non-macOS editor chains
// ---------------------------------------------------------------------------

// UNLIKE the fixtures at the top of this file, the chains below are MODELED, not
// transcribed. They encode VS Code's documented process model on Linux and
// Windows — Chromium re-execs one binary per child role with the role in a
// --type flag — but no pid in them was observed. They must be replaced with real
// captures before any number derived from ide_node_host is published.
//
// They are still worth having now, because the property they pin does not depend
// on the exact argv: on those platforms the editor's child processes carry the
// editor's OWN basename, which used to match the `code` row and produce a
// confident wrong answer. Anything that regresses that is caught here.

// The gap this whole kind exists to close. Same surface as
// TestVSCodeClaudeChatIsAttributedFromExtHostArgv, on Linux: the extension host
// is basename "code", so before kindIDENodeHost it matched the ide-host row, was
// not argv-eligible, and the cli.js path naming the vendor was never read.
func TestLinuxVSCodeExtensionHostIsArgvEligible(t *testing.T) {
	src := fakeSource{
		1000: {Pid: 1000, Ppid: 1001, Name: "agentdetect"},
		1001: {Pid: 1001, Ppid: 1002, Name: "bash"},
		1002: {Pid: 1002, Ppid: 1003, Name: "code", Cmdline: []string{
			"/usr/share/code/code",
			"/home/x/.vscode/extensions/anthropic.claude-code-2.1.220-linux-x64/resources/claude-code/cli.js",
			"--type=utility", "--utility-sub-type=node.mojom.NodeService",
		}},
		1003: {Pid: 1003, Ppid: 1, Name: "code"},
	}
	res := detect(t, src, 1001, map[string]string{"CLAUDECODE": "1"})

	a := res.Signals.AgentAncestor
	if a == nil {
		t.Fatal("no agent ancestor; a Linux ext host is basename `code` and its argv must still be read")
	}
	if a.Vendor != "claude-code" {
		t.Errorf("vendor = %q, want claude-code", a.Vendor)
	}
	if got := a.MatchedOn(); got != "argv" {
		t.Errorf("matched_on = %q, want argv", got)
	}
	// The weaker Via, not extension_host: the role genuinely is not resolvable
	// here, and reporting the mac-strength value would be a fabricated claim.
	s := res.Signals.IDESpawn
	if s == nil || s.Via != spawnIDENodeHost {
		t.Fatalf("ide_spawn = %+v, want via %q", s, spawnIDENodeHost)
	}
	if s.Vendor != "vscode" {
		t.Errorf("ide_spawn vendor = %q, want vscode — the `code` row still supplies it", s.Vendor)
	}
	// The main editor process is still the IDE host; only the child was refined.
	if res.Signals.IDEHost == nil || res.Signals.IDEHost.Vendor != "vscode" {
		t.Errorf("ide_host = %+v, want vscode", res.Signals.IDEHost)
	}
	// "sae", not "she": shape is appended after attribution, so a node host whose
	// argv named a vendor records what it turned out to be. h survives only where
	// the role stayed unresolved — see the human case below.
	if got := res.Signals.ChainShape; got != "sae" {
		t.Errorf("chain shape = %q, want %q", got, "sae")
	}
}

// The honest limit, stated as a test so it cannot be quietly forgotten: on Linux
// the human integrated terminal produces the SAME ide_node_host that the agent
// surface above does. This is the negative control from
// TestHumanIntegratedTerminalIsDistinguishableFromExtHost, and off macOS it is
// no longer distinguishable. The test asserts the ambiguity rather than papering
// over it — if someone later "fixes" this to report ide_utility, they will have
// invented a discriminator that does not exist.
func TestLinuxIntegratedTerminalIsNotDistinguishableFromExtHost(t *testing.T) {
	src := fakeSource{
		1000: {Pid: 1000, Ppid: 1002, Name: "bash"},
		1002: {Pid: 1002, Ppid: 1003, Name: "code", Cmdline: []string{
			"/usr/share/code/code", "--type=utility", "--utility-sub-type=node.mojom.NodeService",
		}},
		1003: {Pid: 1003, Ppid: 1, Name: "code"},
	}
	res := detect(t, src, 1000, nil)

	if res.Signals.AgentAncestor != nil {
		t.Errorf("ancestor = %+v, want nil", res.Signals.AgentAncestor)
	}
	s := res.Signals.IDESpawn
	if s == nil || s.Via != spawnIDENodeHost {
		t.Fatalf("ide_spawn = %+v, want via %q — the pty host is indistinguishable here", s, spawnIDENodeHost)
	}
	// And it lands in the miss count, which is the documented looseness: a human
	// is in the ide-node-host population by construction on this platform.
	if len(res.Signals.Unattributed) != 1 || res.Signals.Unattributed[0].Kind != kindIDENodeHost {
		t.Errorf("unattributed = %+v, want one ide-node-host", res.Signals.Unattributed)
	}
}

// Windows reaches the same classification as Linux through two extra pieces of
// normalization, both of which have to hold or the whole platform silently
// reports nothing: `Code.exe` resolves to the `code` row, and the backslash
// install path still matches a forward-slash pattern.
//
// Names here are pre-normalized because fakeSource returns ProcInfo verbatim —
// normalizeName runs in the ProcSource, per the ProcInfo.Name contract. The
// basename half is asserted directly below; the argv half is what this exercises.
func TestWindowsVSCodeExtensionHostIsArgvEligible(t *testing.T) {
	if got := normalizeName(`C:\Users\x\AppData\Local\Programs\Microsoft VS Code\Code.exe`); got != "code" {
		t.Fatalf("normalizeName = %q, want %q — the rest of this test assumes it", got, "code")
	}

	src := fakeSource{
		1000: {Pid: 1000, Ppid: 1001, Name: "code", Cmdline: []string{
			`C:\Users\x\AppData\Local\Programs\Microsoft VS Code\Code.exe`,
			`C:\Users\x\.vscode\extensions\anthropic.claude-code-2.1.220-win32-x64\resources\claude-code\cli.js`,
			"--type=utility", "--utility-sub-type=node.mojom.NodeService",
		}},
		1001: {Pid: 1001, Ppid: 1, Name: "code"},
	}
	res := detect(t, src, 1000, nil)

	a := res.Signals.AgentAncestor
	if a == nil || a.Vendor != "claude-code" {
		t.Fatalf("ancestor = %+v, want claude-code", a)
	}
	if s := res.Signals.IDESpawn; s == nil || s.Via != spawnIDENodeHost {
		t.Errorf("ide_spawn = %+v, want via %q", s, spawnIDENodeHost)
	}
}

// The renderer-class children must NOT become node hosts. They never host an
// agent, so classifying them as one would put pure GUI noise into the miss count
// and into the argv-read population.
func TestChromiumRendererRolesAreOrdinaryUtilities(t *testing.T) {
	for _, role := range []string{"renderer", "gpu-process", "zygote", "broker", "crashpad-handler"} {
		src := fakeSource{
			1000: {Pid: 1000, Ppid: 1001, Name: "bash"},
			1001: {Pid: 1001, Ppid: 1002, Name: "code", Cmdline: []string{
				"/usr/share/code/code", "--type=" + role,
			}},
			1002: {Pid: 1002, Ppid: 1, Name: "code"},
		}
		res := detect(t, src, 1000, nil)
		if s := res.Signals.IDESpawn; s == nil || s.Via != spawnIDEUtility {
			t.Errorf("--type=%s: ide_spawn = %+v, want via %q", role, s, spawnIDEUtility)
		}
		if len(res.Signals.Unattributed) != 0 {
			t.Errorf("--type=%s: unattributed = %+v, want none — a renderer is not a candidate host",
				role, res.Signals.Unattributed)
		}
	}
}

// An unrecognized --type value classifies nothing, so the process keeps the kind
// its basename gave it. Same allow-list posture as argvEligible: a role we have
// not reasoned about must not acquire a meaning by default.
func TestUnrecognizedChromiumRoleDoesNotReclassify(t *testing.T) {
	src := fakeSource{
		1000: {Pid: 1000, Ppid: 1001, Name: "bash"},
		1001: {Pid: 1001, Ppid: 1, Name: "code", Cmdline: []string{
			"/usr/share/code/code", "--type=some-future-role",
		}},
	}
	res := detect(t, src, 1000, nil)
	if res.Signals.IDESpawn != nil {
		t.Errorf("ide_spawn = %+v, want nil", res.Signals.IDESpawn)
	}
	if res.Signals.IDEHost == nil || res.Signals.IDEHost.Vendor != "vscode" {
		t.Errorf("ide_host = %+v, want vscode", res.Signals.IDEHost)
	}
}

// The main editor process has no --type at all and must stay the IDE host — the
// role rule only ever refines a child, never reinterprets the parent.
func TestEditorMainProcessIsUnaffectedByRoleRule(t *testing.T) {
	src := fakeSource{
		1000: {Pid: 1000, Ppid: 1001, Name: "zsh"},
		1001: {Pid: 1001, Ppid: 1, Name: "code", Cmdline: []string{"/usr/share/code/code", "."}},
	}
	res := detect(t, src, 1000, nil)
	if res.Signals.IDESpawn != nil {
		t.Errorf("ide_spawn = %+v, want nil", res.Signals.IDESpawn)
	}
	if got := res.Signals.ChainShape; got != "se" {
		t.Errorf("chain shape = %q, want %q", got, "se")
	}
}

// macOS must keep its stronger classification. The role rule runs only for
// kindIDEHost and kindUnknown, so a helper whose basename already resolved the
// role is never revisited — even though its argv is the same NodeService utility
// line that yields the weaker kind on Linux. This is the regression that would
// silently flatten the one platform where the split is real.
func TestMacHelperRolesSurviveTheChromiumRoleRule(t *testing.T) {
	extHost := fakeSource{
		1000: {Pid: 1000, Ppid: 1001, Name: "zsh"},
		1001: {Pid: 1001, Ppid: 1002, Name: "code helper (plugin)", Cmdline: []string{
			"/Applications/Visual Studio Code.app/Contents/Frameworks/Code Helper (Plugin).app/Contents/MacOS/Code Helper (Plugin)",
			"--type=utility", "--utility-sub-type=node.mojom.NodeService",
		}},
		1002: {Pid: 1002, Ppid: 1, Name: "code"},
	}
	if s := detect(t, extHost, 1000, nil).Signals.IDESpawn; s == nil || s.Via != spawnExtensionHost {
		t.Errorf("ide_spawn = %+v, want via %q — the bundle name resolves the role on darwin", s, spawnExtensionHost)
	}

	ptyHost := fakeSource{
		1000: {Pid: 1000, Ppid: 1001, Name: "zsh"},
		1001: {Pid: 1001, Ppid: 1002, Name: "code helper", Cmdline: []string{
			"/Applications/Visual Studio Code.app/Contents/Frameworks/Code Helper.app/Contents/MacOS/Code Helper",
			"--type=utility", "--utility-sub-type=node.mojom.NodeService",
		}},
		1002: {Pid: 1002, Ppid: 1, Name: "code"},
	}
	if s := detect(t, ptyHost, 1000, nil).Signals.IDESpawn; s == nil || s.Via != spawnIDEUtility {
		t.Errorf("ide_spawn = %+v, want via %q", s, spawnIDEUtility)
	}
}

// A fork we have no editor row for still yields the role, with an empty vendor —
// the same partial-hit behavior TestHelperClassificationGeneralizesAcrossForks
// pins for macOS. Without this the two platforms would degrade differently for
// the same editor.
func TestLinuxForkYieldsRoleWithoutVendor(t *testing.T) {
	src := fakeSource{
		1000: {Pid: 1000, Ppid: 1001, Name: "bash"},
		1001: {Pid: 1001, Ppid: 1, Name: "someeditor", Cmdline: []string{
			"/opt/someeditor/someeditor", "--type=utility", "--utility-sub-type=node.mojom.NodeService",
		}},
	}
	res := detect(t, src, 1000, nil)
	s := res.Signals.IDESpawn
	if s == nil || s.Via != spawnIDENodeHost {
		t.Fatalf("ide_spawn = %+v, want via %q", s, spawnIDENodeHost)
	}
	if s.Vendor != "" {
		t.Errorf("vendor = %q, want empty — role is known, vendor is not", s.Vendor)
	}
}

// --type= is a CHROMIUM flag, not a VS Code one, so an Electron app we cannot
// identify emits it too. Refining an unidentified product by the renderer-class
// roles would invent a signal, and one of them — ide_utility — is positive
// evidence for a HUMAN, so the invention runs in the direction that quietly
// deflates the agent share.
//
// These pin the gate in classifyChromiumRole: the roles that carry recall are
// still accepted for an unknown product, the ones that would invent are not, and
// a known editor is unaffected.
func TestUnknownElectronAppDoesNotAcquireAnEditorRole(t *testing.T) {
	// Every role we decline to infer without an editor row behind it.
	for _, role := range []string{"renderer", "gpu-process", "zygote", "broker", "crashpad-handler", "utility"} {
		src := fakeSource{
			1000: {Pid: 1000, Ppid: 1001, Name: "bash"},
			1001: {Pid: 1001, Ppid: 1, Name: "slack", Cmdline: []string{
				"/usr/lib/slack/slack", "--type=" + role,
			}},
		}
		res := detect(t, src, 1000, nil)

		if s := res.Signals.IDESpawn; s != nil {
			t.Errorf("--type=%s on an unidentified app: ide_spawn = %+v, want none", role, s)
		}
		// Still unknown, therefore still argv-eligible and still counted. The
		// refinement would have taken both of those away.
		if got := res.Signals.ChainShape; got != "s?" {
			t.Errorf("--type=%s: chain_shape = %q, want %q", role, got, "s?")
		}
		if len(res.Signals.Unattributed) != 1 {
			t.Errorf("--type=%s: unattributed = %+v, want the app counted once",
				role, res.Signals.Unattributed)
		}
	}
}

// The other half of the gate. These two roles are what rescue a VS Code fork we
// have no row for, and neither can claim a human: an extension host is a
// stronger agent-side claim than unknown, and a node host is explicitly
// two-population. TestLinuxForkYieldsRoleWithoutVendor covers the NodeService
// case; this is the named extension host.
func TestUnknownForkStillYieldsAnExtensionHost(t *testing.T) {
	src := fakeSource{
		1000: {Pid: 1000, Ppid: 1001, Name: "bash"},
		1001: {Pid: 1001, Ppid: 1, Name: "someeditor", Cmdline: []string{
			"/opt/someeditor/someeditor", "--type=extensionHost",
		}},
	}
	res := detect(t, src, 1000, nil)
	if s := res.Signals.IDESpawn; s == nil || s.Via != spawnExtensionHost {
		t.Fatalf("ide_spawn = %+v, want via %q — this is the recall the gate must not cost",
			s, spawnExtensionHost)
	}
}

// Positive control: with an editor row behind it, every role is credible again.
func TestKnownEditorKeepsTheFullRefinement(t *testing.T) {
	for _, role := range []string{"renderer", "ptyhost", "gpu-process"} {
		src := fakeSource{
			1000: {Pid: 1000, Ppid: 1001, Name: "bash"},
			1001: {Pid: 1001, Ppid: 1002, Name: "code", Cmdline: []string{
				"/usr/share/code/code", "--type=" + role,
			}},
			1002: {Pid: 1002, Ppid: 1, Name: "code"},
		}
		res := detect(t, src, 1000, nil)

		s := res.Signals.IDESpawn
		if s == nil || s.Via != spawnIDEUtility {
			t.Errorf("--type=%s under a known editor: ide_spawn = %+v, want via %q",
				role, s, spawnIDEUtility)
		}
		if s != nil && s.Vendor != "vscode" {
			t.Errorf("--type=%s: vendor = %q, want vscode from the `code` row", role, s.Vendor)
		}
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
