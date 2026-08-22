package agentdetect

import "strings"

// Fingerprint tables.
//
// Intended to ship via feature flag as a JSON document with a compiled-in
// fallback, so they can iterate independently of CLI releases. A starting point
// from field tests and public agent packaging conventions — NOT verified against
// every vendor's current release, and may be wrong now or later.

// ---------------------------------------------------------------------------
// Signal: environment variables
// ---------------------------------------------------------------------------

// envFingerprints lists env vars an agent vendor is known to set.
// AI_AGENT and AGENT are the proposed-standard variables, but many vendors have their own.
var envFingerprints = []string{
	"CLAUDECODE",
	"CLAUDE_CODE",
	"CURSOR_AGENT",
	"CODEX_SANDBOX",
	"CODEX_CI",
	"CODEX_THREAD_ID",
	"GEMINI_CLI",
	"COPILOT_CLI",
	"COPILOT_MODEL",
	"COPILOT_AGENT",
	"REPL_ID",
	"AI_AGENT",
	"AGENT",
}

// vendorForProcKey and vendorForArgvPattern resolve an ancestry key back to a
// vendor — a basename or argv pattern doesn't always spell one out ("q" is
// amazon-q), so this is the reference mapping. Two lookups because the key spaces
// are distinct ("code" vs "@openai/codex"), as the wire format keeps them.
func vendorForProcKey(key string) (string, bool) {
	fp, ok := procFingerprints[key]
	if !ok {
		return "", false
	}
	return fp.Vendor, true
}

func vendorForArgvPattern(pattern string) (string, bool) {
	for _, fp := range cmdlineFingerprints {
		if fp.Pattern == pattern {
			return fp.Vendor, true
		}
	}
	return "", false
}

// CI variables — not agent signals, but context: an agent var inside CI is more
// likely an inherited image setting than a live agent.
type ciFingerprint struct {
	Var      string
	Provider string
}

var ciFingerprints = []ciFingerprint{
	{Var: "GITHUB_ACTIONS", Provider: "github-actions"},
	{Var: "GITLAB_CI", Provider: "gitlab-ci"},
	{Var: "BUILDKITE", Provider: "buildkite"},
	{Var: "CIRCLECI", Provider: "circleci"},
	{Var: "JENKINS_URL", Provider: "jenkins"},
	{Var: "SEMAPHORE", Provider: "semaphore"},
	{Var: "TF_BUILD", Provider: "azure-pipelines"},
}

// genericCIVar is set by essentially every CI system, including all of the above,
// so on its own without a vendor match it means "some CI we cannot name".
const genericCIVar = "CI"

// ---------------------------------------------------------------------------
// Signal: executable basename
// ---------------------------------------------------------------------------

type procKind string

const (
	kindAgent   procKind = "agent"
	kindIDEHost procKind = "ide-host" // editor that may host an agent, but isn't one

	kindInterpreter procKind = "interpreter" // name says nothing; identity is in argv
	kindShell       procKind = "shell"       // expected in the chain; keep walking
	kindTerminal    procKind = "terminal"    // terminal emulator; usually the human boundary
	kindWrapper     procKind = "wrapper"     // make/npm/etc; keep walking
	kindRemote      procKind = "remote"      // sshd; ancestry is severed above this
	kindInit        procKind = "init"
	kindUnknown     procKind = "unknown"
)

// code is the single character used in Signals.ChainShape.
func (k procKind) code() byte {
	switch k {
	case kindAgent:
		return 'a'
	case kindIDEHost:
		return 'e'
	case kindInterpreter:
		return 'i'
	case kindShell:
		return 's'
	case kindTerminal:
		return 't'
	case kindWrapper:
		return 'w'
	case kindRemote:
		return 'r'
	case kindInit:
		return 'n'
	default:
		return '?'
	}
}

type procFingerprint struct {
	Vendor string
	Kind   procKind
}

// Some keys below are weak identities — "q" (Amazon Q, but also a one-letter
// binary), "amp", "code" (VS Code, but also unrelated tools) and "studio".
var procFingerprints = map[string]procFingerprint{
	// Agents that ship as a named executable — the happy path.
	"claude":       {Vendor: "claude-code", Kind: kindAgent},
	"cursor-agent": {Vendor: "cursor", Kind: kindAgent},
	"codex":        {Vendor: "codex", Kind: kindAgent},
	"gemini":       {Vendor: "gemini-cli", Kind: kindAgent},
	"copilot":      {Vendor: "github-copilot", Kind: kindAgent},
	"aider":        {Vendor: "aider", Kind: kindAgent},
	"goose":        {Vendor: "goose", Kind: kindAgent},
	"opencode":     {Vendor: "opencode", Kind: kindAgent},
	"cline":        {Vendor: "cline", Kind: kindAgent},
	"crush":        {Vendor: "crush", Kind: kindAgent},
	"amp":          {Vendor: "amp", Kind: kindAgent},
	"q":            {Vendor: "amazon-q", Kind: kindAgent},

	// Editors that host agents. Presence means "agent-capable environment",
	// not "agent-initiated call".
	"cursor":   {Vendor: "cursor", Kind: kindIDEHost},
	"code":     {Vendor: "vscode", Kind: kindIDEHost},
	"windsurf": {Vendor: "windsurf", Kind: kindIDEHost},
	"zed":      {Vendor: "zed", Kind: kindIDEHost},

	// JetBrains ships one launcher per product; "jetbrains" is never a basename.
	// The version-suffix rule below also covers idea64/rider64 (Windows bitness).
	"idea":     {Vendor: "jetbrains", Kind: kindIDEHost},
	"pycharm":  {Vendor: "jetbrains", Kind: kindIDEHost},
	"goland":   {Vendor: "jetbrains", Kind: kindIDEHost},
	"webstorm": {Vendor: "jetbrains", Kind: kindIDEHost},
	"phpstorm": {Vendor: "jetbrains", Kind: kindIDEHost},
	"rubymine": {Vendor: "jetbrains", Kind: kindIDEHost},
	"clion":    {Vendor: "jetbrains", Kind: kindIDEHost},
	"rider":    {Vendor: "jetbrains", Kind: kindIDEHost},
	"datagrip": {Vendor: "jetbrains", Kind: kindIDEHost},
	// Android Studio is IntelliJ-derived; "studio" is a weak identity (see above).
	"studio": {Vendor: "jetbrains", Kind: kindIDEHost},

	// Interpreters: the name identifies nothing, so argv is the only evidence.
	"node":   {Kind: kindInterpreter},
	"bun":    {Kind: kindInterpreter},
	"deno":   {Kind: kindInterpreter},
	"python": {Kind: kindInterpreter},
	"ruby":   {Kind: kindInterpreter},
	"java":   {Kind: kindInterpreter},
	// Versioned basenames ("python3.13") resolve to the stem via the rule below.
	// uv/uvx: a common install path for Python agents (e.g. aider). Interpreter,
	// not wrapper, because the program is named in argv (`uv tool run aider`) — a
	// wrapper kind would exclude that from matching.
	"uv": {Kind: kindInterpreter}, "uvx": {Kind: kindInterpreter},

	// Expected intermediaries. Keep walking.
	"sh": {Kind: kindShell}, "bash": {Kind: kindShell}, "zsh": {Kind: kindShell},
	"fish": {Kind: kindShell}, "dash": {Kind: kindShell}, "ksh": {Kind: kindShell},
	"pwsh": {Kind: kindShell}, "powershell": {Kind: kindShell}, "cmd": {Kind: kindShell},
	"login": {Kind: kindShell},

	"make": {Kind: kindWrapper}, "npm": {Kind: kindWrapper}, "npx": {Kind: kindWrapper},
	"yarn": {Kind: kindWrapper}, "pnpm": {Kind: kindWrapper}, "task": {Kind: kindWrapper},
	"just": {Kind: kindWrapper}, "xargs": {Kind: kindWrapper}, "timeout": {Kind: kindWrapper},
	"env": {Kind: kindWrapper}, "sandbox-exec": {Kind: kindWrapper}, "sudo": {Kind: kindWrapper},

	// Cursor's sandbox helper, applied to agent commands but not the integrated
	// terminal — so its presence is itself corroborating evidence.
	"cursorsandbox": {Vendor: "cursor", Kind: kindWrapper},

	"alacritty": {Kind: kindTerminal}, "iterm2": {Kind: kindTerminal},
	"terminal": {Kind: kindTerminal}, "kitty": {Kind: kindTerminal},
	"wezterm": {Kind: kindTerminal}, "wezterm-gui": {Kind: kindTerminal},
	"ghostty": {Kind: kindTerminal}, "tmux": {Kind: kindTerminal},
	"screen": {Kind: kindTerminal}, "windowsterminal": {Kind: kindTerminal},

	// Above this point ancestry tells us nothing about the caller.
	"sshd": {Kind: kindRemote},
	"init": {Kind: kindInit}, "systemd": {Kind: kindInit}, "launchd": {Kind: kindInit},
}

// ---------------------------------------------------------------------------
// Signal: editor helper processes, matched by the Electron " helper" suffix
// ---------------------------------------------------------------------------

// Electron names child processes "<Product> Helper[ (Role)]". Matching the suffix
// maps any role of a known editor's helper to that editor; classifyEditorHelper
// gates it on a known editor so non-editor Electron apps aren't swept in.
const helperSuffix = " helper"

// resolveFingerprint identifies a process by, in order, its executable basename,
// argv[0]'s basename, then the directories the executable sits in — each tried
// only if the previous missed, and each returning a table KEY, never an observed
// string.
//
// The path fallback matters because Claude Code's native install names the binary
// by version (.local/share/claude/versions/2.1.231): the basename is a bare
// version and argv[0] a bare "claude", so without it the install is never
// attributed by ancestry.
func resolveFingerprint(info ProcInfo) (procFingerprint, string, bool) {
	if fp, key, ok := lookupFingerprint(normalizeName(info.Name)); ok {
		return fp, key, true
	}

	// argv[0] is the program's own identity claim, not user-authored command
	// text — a shell's argv[0] is the shell, not what it was asked to run.
	//
	// Not authoritative enough to come first: it's settable (exec -a), and a
	// login shell presents as "-zsh", matching nothing. Only consulted once the
	// real basename has already failed.
	if len(info.Cmdline) > 0 {
		if fp, key, ok := lookupFingerprint(normalizeName(info.Cmdline[0])); ok {
			return fp, key, true
		}
	}

	return lookupVersionedPath(info.Name)
}

// versionSegmentChars are what a path segment made of nothing but a version is
// built from. A leading "v" is tolerated: "v20.11.0".
const versionSegmentChars = "0123456789."

// maxPathSegmentHops is how far above the executable to look: two reaches
// <product>/versions/<ver> and stops short of a home dir or checkout root.
const maxPathSegmentHops = 2

// lookupVersionedPath resolves a program whose basename is a version by reading
// the directories it sits in.
//
// Triggered only by a version-shaped basename — a rule that searched every
// unrecognized binary's path would attribute an agent to anyone with a checkout
// at ~/dev/cursor. Restricted to argv-eligible kinds so it can only ever add an
// attribution: typing a kindIDEHost/kindWrapper process would strip its argv
// eligibility and turn a counted miss into an invisible one.
func lookupVersionedPath(path string) (procFingerprint, string, bool) {
	segments := strings.FieldsFunc(strings.ToLower(strings.TrimSpace(path)), func(r rune) bool {
		return r == '/' || r == '\\'
	})
	if len(segments) < 2 {
		return procFingerprint{}, "", false
	}

	base := strings.TrimSuffix(segments[len(segments)-1], ".exe")
	if !isVersionSegment(base) {
		return procFingerprint{}, "", false
	}

	for hop := 1; hop <= maxPathSegmentHops; hop++ {
		i := len(segments) - 1 - hop
		if i < 0 {
			break
		}
		fp, ok := procFingerprints[segments[i]]
		if !ok || !argvEligible(fp.Kind) {
			continue
		}
		return fp, segments[i], true
	}
	return procFingerprint{}, "", false
}

// isVersionSegment reports whether a path segment is nothing but a version.
func isVersionSegment(s string) bool {
	s = strings.TrimPrefix(s, "v")
	if s == "" || strings.Trim(s, versionSegmentChars) != "" {
		return false
	}
	return strings.ContainsAny(s, "0123456789")
}

// lookupFingerprint resolves a normalized basename, preferring the exact table,
// then the version-suffix rule, then the Electron helper rule. key is the matched
// table key, or "" when the helper rule matched — only a key is licensed to be
// recorded, which is also why a versioned name resolves to its stem.
func lookupFingerprint(name string) (procFingerprint, string, bool) {
	if fp, ok := procFingerprints[name]; ok {
		return fp, name, true
	}
	if fp, key, ok := lookupVersionedName(name); ok {
		return fp, key, true
	}
	fp, ok := classifyEditorHelper(name)
	return fp, "", ok
}

// versionSuffixChars are the characters a trailing version or bitness suffix is
// built from: "python3.13", "idea64".
const versionSuffixChars = "0123456789."

// lookupVersionedName resolves a versioned basename to its table stem
// ("python3.13" → "python", "idea64" → "idea"), so new versions aren't silent
// misses. Restricted to interpreters and editor hosts — the kinds whose basenames
// carry versions, neither an agent attribution; trimming digits off an unknown
// binary until it hit an agent row would fabricate a vendor.
func lookupVersionedName(name string) (procFingerprint, string, bool) {
	stem := strings.TrimRight(name, versionSuffixChars)
	if stem == name || stem == "" {
		return procFingerprint{}, "", false
	}
	fp, ok := procFingerprints[stem]
	if !ok || (fp.Kind != kindInterpreter && fp.Kind != kindIDEHost) {
		return procFingerprint{}, "", false
	}
	return fp, stem, true
}

func classifyEditorHelper(name string) (procFingerprint, bool) {
	// "code helper (plugin)" → "code"; the role is discarded — a helper only
	// means "we're inside that editor".
	product, ok := splitHelperName(name)
	if !ok {
		return procFingerprint{}, false
	}

	// Gate on a KNOWN editor. The " helper" suffix is Electron's, not any
	// editor's, so Slack, Discord, Hyper (a terminal) and every other Electron
	// app match the shape too — classifying those as IDEHost would pollute the
	// signal. An unrecognized product falls through to kindUnknown instead.
	fp, known := procFingerprints[product]
	if !known || fp.Kind != kindIDEHost {
		return procFingerprint{}, false
	}
	return procFingerprint{Vendor: fp.Vendor, Kind: kindIDEHost}, true
}

// splitHelperName returns the product from "<product> helper[ (role)]", stripping
// any "(role)". A bare "helper" (empty product) is not a match.
func splitHelperName(name string) (string, bool) {
	rest := name
	if i := strings.IndexByte(rest, '('); i >= 0 {
		rest = strings.TrimSpace(rest[:i])
	}
	product, found := strings.CutSuffix(rest, helperSuffix)
	if !found || product == "" {
		return "", false
	}
	return product, true
}

// ---------------------------------------------------------------------------
// Signal: command line, for agents whose executable name says nothing
// ---------------------------------------------------------------------------

// Matched as a lowercased substring against the identity POSITIONS of an
// argv-eligible ancestor (see argvIdentityFields); first match wins. This table
// carries most of the recall, since agents are usually node scripts.
//
// Safe because matching is against THIS fixed list only (no heuristics); only the
// Pattern and Vendor are emitted, never raw argv (see AncestorMatch); and
// argvEligible / argvIdentityFields keep matching off user-authored command text.
//
// Patterns must be specific enough to be identities — "@anthropic-ai/claude-code"
// can't occur by accident, a bare "claude" would match shell history. Editor
// extensions with no CLI (cody, continue, tabby) are absent: no argv to match.
type cmdlineFingerprint struct {
	Pattern string
	Vendor  string
}

var cmdlineFingerprints = []cmdlineFingerprint{
	{Pattern: "@anthropic-ai/claude-code", Vendor: "claude-code"},
	{Pattern: "claude-code/cli.js", Vendor: "claude-code"},
	{Pattern: "/claude/cli.js", Vendor: "claude-code"},
	{Pattern: "@openai/codex", Vendor: "codex"},
	{Pattern: "@google/gemini-cli", Vendor: "gemini-cli"},
	{Pattern: "@github/copilot", Vendor: "github-copilot"},
	{Pattern: "cursor-agent", Vendor: "cursor"},
	{Pattern: "opencode-ai", Vendor: "opencode"},
	{Pattern: "site-packages/aider", Vendor: "aider"},
	{Pattern: "/aider/", Vendor: "aider"},
	{Pattern: "goose-ai", Vendor: "goose"},
}
