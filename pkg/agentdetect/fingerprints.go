package agentdetect

import "strings"

// Fingerprint tables.
//
// In production tables should ship as a LaunchDarkly-delivered JSON
// document w/compiled-in fallback version, for easier iterating.
//
// Everything below is a starting point assembled from research, a proof of
// concept, and public agent packaging conventions. It is NOT verified against
// every vendor's current release and may be wrong now or in the future.

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

// TODO NC - do we care about Devin? Where are we looking for this path?
// fileFingerprints is for agents that identify by presence of a path rather than an env var
var fileFingerprints = []string{
	"/opt/.devin",
}

// filePrefix namespaces a filesystem marker inside Signals.AgentEnv, so one
// list can carry both env vars and file markers ("same signal, different
// evidence location") without a near-empty second field per class.
//
// The reservation is structural, not by convention: no environment variable
// name can contain a colon, so a key either starts with a prefix or is a
// variable, with no overlap possible. TestEnvKeysAreDistinguishable pins it.
//
// A future marker class (registry key, socket, mount) takes a new const here,
// with no change to Signals.AgentEnv or existing rows. Keep prefixes short,
// lowercase, and colon-terminated.
const filePrefix = "file:"

// vendorForProcKey and vendorForArgvPattern map the two ancestry key
// populations back to a vendor: Attributes.AgentProc resolves through the
// first, Attributes.AgentArgv through the second. Unlike the environment
// markers above, a basename or argv pattern doesn't always spell out its
// vendor ("q" is Amazon Q, "@openai/codex" is codex), so the mapping is kept
// here as the reference implementation analytics owns at query time — a row
// added without a vendor fails a test instead of silently producing an
// unmappable key.
//
// Two lookups rather than one because the key spaces are genuinely distinct —
// "code" vs. "@openai/codex" — matching how the wire format keeps them in
// separate fields.
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

// CI variables. Not agent signals, but they change how an agent signal should
// be read — an agent var inside CI is more likely an inherited image setting
// than a live agent.
//
// Provider is what gets reported, not Var: a normalized id, same discipline
// as the vendor ids.
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
// so on its own it means "some CI we cannot name" and folded into the provider
// list it would be a near-constant member. Reported separately for that reason.
const genericCIVar = "CI"

// ---------------------------------------------------------------------------
// Signal: executable basename
// ---------------------------------------------------------------------------

type procKind string

const (
	kindAgent   procKind = "agent"
	kindIDEHost procKind = "ide-host" // editor that may host an agent, but isn't one

	// kindIDEExtHost is an editor's EXTENSION HOST, or a process it spawned. A
	// stronger claim than kindIDEHost: an extension host in the ancestry means
	// an extension — not the integrated terminal — spawned this command.
	//
	// Observed on darwin/arm64: Claude Code's VS Code extension, GitHub Copilot
	// chat, and Cursor's agent all route tool calls through a "* Helper
	// (Plugin)" process, while the integrated terminal routes through the pty
	// host, a NON-plugin helper (kindIDEUtility below).
	//
	// Honest limit: this means extension-initiated, not strictly
	// agent-initiated — a task-runner extension or a human clicking "run" in an
	// extension's UI lands here too. Still far tighter than "an editor is
	// somewhere above us"; the finer distinction is left to downstream
	// analytics.
	//
	// Also argv-eligible: these processes are Electron running as node
	// (ELECTRON_RUN_AS_NODE), so an agent shipped as a node script has its real
	// identity in argv even though its basename is the helper's.
	kindIDEExtHost procKind = "ide-ext-host"

	// kindIDEUtility is any other editor helper process — pty host, renderer,
	// GPU, file watcher, shared process. The one that matters is the pty host:
	// seeing it in a shell ancestry is positive evidence of a human-typed
	// command.
	kindIDEUtility procKind = "ide-utility"

	// kindIDENodeHost is an editor utility process hosting node, where we can
	// tell that much and no more. Exists because outside macOS, kindIDEExtHost
	// and kindIDEUtility are not separable.
	//
	// On macOS the role is in the basename (extension host runs from the
	// "(Plugin)" bundle, pty host from the plain one), even though their argv is
	// byte-identical (`--type=utility --utility-sub-type=node.mojom.NodeService`;
	// see fixtures in ide_surfaces_test.go). Linux and Windows have no per-role
	// bundles — Chromium re-execs the same `code` / `Code.exe` binary and the
	// role lives only in that identical argv.
	//
	// So on those platforms a NodeService utility parent is EITHER an agent's
	// extension host or a human's integrated terminal, and no signal we collect
	// resolves it. This is still strictly better than treating it as kindIDEHost
	// (not argv-eligible, an invisible miss): argv is read here, so an agent
	// naming itself is attributed and one that doesn't is at least counted in
	// Unattributed.
	//
	// Deliberately separate from kindIDEExtHost rather than reused, since that
	// would report a mac-strength claim from a Linux-strength observation and
	// turn the integrated-terminal negative control into a false positive.
	kindIDENodeHost procKind = "ide-node-host"

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
	case kindIDEExtHost:
		return 'x'
	case kindIDEUtility:
		return 'u'
	case kindIDENodeHost:
		return 'h'
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
// There's no Ambiguous flag for these: the payload carries the matched KEY,
// not the vendor, so ambiguity is a property analytics can already derive
// from the key itself.

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
	"devin":        {Vendor: "devin", Kind: kindAgent},
	"crush":        {Vendor: "crush", Kind: kindAgent},
	"amp":          {Vendor: "amp", Kind: kindAgent},
	"q":            {Vendor: "amazon-q", Kind: kindAgent},

	// Editors that host agents. Presence means "agent-capable environment",
	// not "agent-initiated call" — a human using the built-in terminal looks
	// identical. Reported separately for exactly that reason.
	"cursor":   {Vendor: "cursor", Kind: kindIDEHost},
	"code":     {Vendor: "vscode", Kind: kindIDEHost},
	"windsurf": {Vendor: "windsurf", Kind: kindIDEHost},
	"zed":      {Vendor: "zed", Kind: kindIDEHost},

	// JetBrains ships one launcher per product; "jetbrains" is never a process
	// basename. The version-suffix rule below strips Windows' "64" bitness
	// suffix, so these rows also cover idea64/rider64/... without a second row
	// per product.
	"idea":     {Vendor: "jetbrains", Kind: kindIDEHost},
	"pycharm":  {Vendor: "jetbrains", Kind: kindIDEHost},
	"goland":   {Vendor: "jetbrains", Kind: kindIDEHost},
	"webstorm": {Vendor: "jetbrains", Kind: kindIDEHost},
	"phpstorm": {Vendor: "jetbrains", Kind: kindIDEHost}, "rubymine": {Vendor: "jetbrains", Kind: kindIDEHost},
	"clion": {Vendor: "jetbrains", Kind: kindIDEHost}, "rider": {Vendor: "jetbrains", Kind: kindIDEHost},
	"datagrip": {Vendor: "jetbrains", Kind: kindIDEHost},
	// Android Studio is IntelliJ-derived. "studio" is a common enough word to be
	// a weak identity; see the note above the procFingerprint type.
	"studio": {Vendor: "jetbrains", Kind: kindIDEHost},

	// Interpreters: the name identifies nothing, so argv is the only evidence.
	"node":   {Kind: kindInterpreter},
	"bun":    {Kind: kindInterpreter},
	"deno":   {Kind: kindInterpreter},
	"python": {Kind: kindInterpreter},
	"ruby":   {Kind: kindInterpreter},
	"java":   {Kind: kindInterpreter},
	// Versioned basenames ("python3.13") aren't rows here — the version-suffix
	// rule below resolves them to this stem.
	// uv/uvx: a mainstream install path for Python-packaged agents (incl.
	// aider). Interpreter rather than wrapper because the program run is named
	// in their argv (`uv tool run aider`), which a wrapper kind would exclude
	// from matching.
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

	// Cursor's sandbox helper, applied to agent-run commands but NOT the
	// integrated terminal — unlike the generic wrappers above, its presence is
	// itself corroborating provenance evidence.
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
// Signal: editor helper processes, classified by shape rather than by name
// ---------------------------------------------------------------------------

// Electron packages its child processes as sibling .app bundles named
// "<Product> Helper[ (Role)]", so the executable basename encodes the role — a
// property of Electron's packaging, not of any one vendor, which is why this
// is a suffix rule rather than table rows.
//
// A rule rather than enumeration means an editor we've never seen is still
// classified correctly (role known, vendor unknown) on first contact, instead
// of being a silent miss until someone notices a new fork.
const (
	helperSuffix    = " helper"
	extHostRoleName = "(plugin)"
)

// resolveFingerprint identifies a process from everything we observed about
// it, in decreasing order of how directly the evidence names the program: the
// executable basename, then argv[0]'s basename, then the directories the
// executable sits in.
//
// The basename alone is not enough: Claude Code's native install is
// .local/share/claude/versions/2.1.231 — the basename is a bare version, and
// argv[0] is a bare "claude" that no pattern should match (a bare "claude"
// pattern would fire on any file of that name). Without the path-segment
// fallback, this install is never attributed by ancestry on any platform —
// it lands as an anonymous entry in Unattributed.
//
// Each step below is tried only when the previous found nothing, and every
// one returns a table KEY, so a match here is still vocabulary, never an
// observed string.
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

// maxPathSegmentHops is how far above the executable to look. Two reaches the
// product directory of the layout that motivated this — <product>/versions/<ver>
// — and stops well short of a home directory or a checkout root.
const maxPathSegmentHops = 2

// lookupVersionedPath resolves a program whose basename is a version by
// reading the directories it sits in.
//
// Deliberately triggered only by a version-shaped basename, not by "the
// basename did not resolve" — a rule that searched every unrecognized
// binary's path would attribute an agent to anyone whose username is
// "claude" or who has a checkout at ~/dev/cursor.
//
// Restricted to argv-eligible kinds, since a path segment is weaker evidence
// than a basename and a wrong guess is asymmetric: typing a process
// kindIDEHost or kindWrapper removes its argv eligibility, converting a
// counted Unattributed entry into an invisible miss. This restriction means
// the rule can only ever add an attribution, never delete a signal.
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

// lookupFingerprint resolves a normalized basename to a fingerprint,
// preferring the exact table, then the version-suffix rule, then the
// Electron helper rule.
//
// key is the table key that matched, or "" when the match came from the
// helper shape rule rather than a row — a key is vocabulary licensed to be
// recorded (see walk), whereas the observed basename is not, which is why a
// versioned name resolves to its STEM here rather than reporting itself.
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

// lookupVersionedName resolves a versioned basename to the table stem it
// belongs to — "python3.13" → "python", "idea64" → "idea". A rule rather
// than table rows, since enumerating means every new version is a silent
// miss until someone notices.
//
// Restricted to interpreters and editor hosts deliberately: those are the
// kinds whose real basenames carry versions, and neither is an agent
// attribution. Trimming digits off an unrecognized binary until it landed on
// an agent row would fabricate a vendor — the same error the argv precision
// guards exist to prevent.
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
	// "code helper (plugin)" → product "code", role "(plugin)".
	// "code helper"          → product "code", role "".
	product, role, ok := splitHelperName(name)
	if !ok {
		return procFingerprint{}, false
	}

	// Vendor comes from the existing editor rows, so "cursor helper (plugin)"
	// attributes to the same vendor id as "cursor" itself. An unrecognized
	// product yields an empty vendor and a known role.
	var vendor string
	if fp, known := procFingerprints[product]; known && fp.Kind == kindIDEHost {
		vendor = fp.Vendor
	}

	kind := kindIDEUtility
	if role == extHostRoleName {
		kind = kindIDEExtHost
	}
	return procFingerprint{Vendor: vendor, Kind: kind}, true
}

// splitHelperName recognizes "<product> helper" and "<product> helper (<role>)".
// The product must be non-empty, so a bare "helper" is not an editor helper.
func splitHelperName(name string) (string, string, bool) {
	rest := name
	var role string
	if i := strings.IndexByte(rest, '('); i >= 0 {
		role = strings.TrimSpace(rest[i:])
		rest = strings.TrimSpace(rest[:i])
	}
	product, found := strings.CutSuffix(rest, helperSuffix)
	if !found || product == "" {
		return "", "", false
	}
	return product, role, true
}

// ---------------------------------------------------------------------------
// Signal: Chromium process role, for the platforms that keep it out of the name
// ---------------------------------------------------------------------------

// classifyEditorHelper above covers macOS, where Electron gives each child
// role its own .app bundle and basename. Linux and Windows have no bundles:
// Chromium re-execs the SAME executable for every child, so VS Code's
// extension host, pty host, renderer and GPU process are all basename `code`
// / `Code.exe`, with the role carried in a `--type=` flag instead. Without
// this, all of them matched the `code` row as kindIDEHost (not
// argv-eligible), making the entire in-editor agent population outside macOS
// an invisible miss.
//
// This is role detection, NOT vendor attribution, which is why it is allowed to
// read the argv of a kind argvEligible otherwise excludes. It matches an exact
// Chromium flag against a fixed value list; it never substring-scans for vendor
// names. The precision hazard argvEligible guards against — argv that is
// user-authored command text — does not exist here, because these flags are
// written by Electron, never by the user.
const (
	chromiumTypeFlag    = "--type"
	chromiumSubTypeFlag = "--utility-sub-type"

	// chromiumUtilityType is the generic utility role. On its own it says nothing
	// about which utility, so it is refined by the sub-type below.
	chromiumUtilityType = "utility"

	// nodeServiceSubType is the utility sub-type Electron uses for a child that
	// runs node. Both the extension host and the pty host are NodeService
	// utilities, which is precisely why this resolves to kindIDENodeHost and not
	// to either of the two specific kinds.
	nodeServiceSubType = "node.mojom.nodeservice"
)

// chromiumRoles is the fixed vocabulary of --type values we act on. An
// unrecognized value classifies nothing — same allow-list posture as
// argvEligible.
//
// extensionhost and ptyhost are here because older VS Code builds fork those
// roles directly with a named type rather than through a utility process, so
// they're unambiguous and get the strong kinds. Everything else is a
// renderer-class child that never hosts an agent.
var chromiumRoles = map[string]procKind{
	"extensionhost":     kindIDEExtHost,
	"ptyhost":           kindIDEUtility,
	"renderer":          kindIDEUtility,
	"gpu-process":       kindIDEUtility,
	"zygote":            kindIDEUtility,
	"broker":            kindIDEUtility,
	"crashpad-handler":  kindIDEUtility,
	chromiumUtilityType: kindIDEUtility,
}

// classifyChromiumRole reads an editor child's role out of its argv. The bool
// is false when argv carries no --type at all, when the value is unrecognized,
// or when the role is one we decline to infer for an unidentified product —
// see knownEditor.
//
// knownEditor says whether the basename already matched an editor row, and
// gates the refinement because `--type=` is a CHROMIUM flag, not a VS Code
// one — Slack, Discord, Postman and every other Electron app emits it, so for
// an unidentified process most of these roles would be invented.
//
// The asymmetry: kindIDEUtility is positive evidence for a HUMAN (the pty
// host is what the integrated terminal hangs off), so typing an unidentified
// Electron renderer as one would manufacture human evidence for a non-editor
// and strip its kindUnknown argv eligibility. The two roles that survive for
// an unknown product (extensionhost, NodeService utility) can't make that
// error — neither claims a human, and both stay argv-eligible.
func classifyChromiumRole(argv []string, knownEditor bool) (procKind, bool) {
	typeVal, ok := chromiumFlagValue(argv, chromiumTypeFlag)
	if !ok {
		return "", false
	}
	kind, known := chromiumRoles[typeVal]
	if !known {
		return "", false
	}
	// A bare utility is a renderer-class child. A NodeService utility is the
	// ambiguous one: extension host and pty host are indistinguishable here.
	if typeVal == chromiumUtilityType {
		if sub, ok := chromiumFlagValue(argv, chromiumSubTypeFlag); ok && sub == nodeServiceSubType {
			kind = kindIDENodeHost
		}
	}
	if !knownEditor && kind != kindIDEExtHost && kind != kindIDENodeHost {
		return "", false
	}
	return kind, true
}

// chromiumFlagValue finds `--flag=value`, and also `--flag value` — Chromium
// only emits the inline form, but accepting both costs one branch and removes a
// silent miss if that ever changes.
func chromiumFlagValue(argv []string, flag string) (string, bool) {
	for i, arg := range argv {
		lower := strings.ToLower(arg)
		if value, found := strings.CutPrefix(lower, flag+"="); found {
			return value, value != ""
		}
		if lower == flag && i+1 < len(argv) {
			return strings.ToLower(argv[i+1]), true
		}
	}
	return "", false
}

// ---------------------------------------------------------------------------
// Signal: command line, for agents whose executable name says nothing
// ---------------------------------------------------------------------------

// Matched as a lowercased substring against the identity POSITIONS of the
// argv of any ancestor argvEligible() admits — argv[0] for an unknown
// basename, argv[0] plus the first few non-flag arguments for a host whose
// arguments name what it is running. See argvIdentityFields. Ordered: first
// match wins.
//
// This table carries most of the recall, since the modal packaging for an
// agent is a node script. The guards this approval rests on:
//
//  1. Matching is against THIS fixed list only — no heuristic, no regex over
//     free text, no "looks like an agent" inference.
//  2. What gets emitted is the Pattern and Vendor, both table entries. Raw
//     argv never enters a payload; see AncestorMatch.
//  3. argvEligible() keeps matching off shells/terminals/wrappers (whose argv
//     is user-authored), and argvIdentityFields() keeps it off the argument
//     positions of everything else.
//
// Deliberately absent: cody, continue, tabby — editor extensions with no CLI
// of their own, so there is no process or argv for a pattern to match.
//
// Patterns should stay specific enough to be identities: a bare "claude"
// would match a filename in shell history, while "@anthropic-ai/claude-code"
// is an installed package path that effectively cannot occur by accident.
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
