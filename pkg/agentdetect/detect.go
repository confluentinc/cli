package agentdetect

import (
	"os"
	"slices"
	"strings"
	"time"

	cliio "github.com/confluentinc/cli/v4/pkg/io"
)

const (
	// Depth and Time budget to guard against a walk that never ends
	defaultMaxDepth = 20
	defaultBudget   = 50 * time.Millisecond
)

// Result is this package's own diagnostic shape, not the wire format used for
// telemetry — see Result.Attributes for that mapping. The JSON tags exist so a
// local `-o json` debug dump is readable; they are not a serialized contract.
type Result struct {
	// Schema versions this struct. Local only: not sent with payload.
	Schema string `json:"schema"`

	// Tables identifies which fingerprint table version produced these
	// Signals. Tables are meant to ship via feature flag independently of CLI
	// releases, so the binary version alone can't identify which vendors were
	// listed when this data was generated. See Attributes.Tables.
	Tables string `json:"tables"`

	Signals  Signals  `json:"signals"`
	Walk     WalkMeta `json:"walk"`
	TimingUs int64    `json:"timing_us"`
}

// Signals holds the independently-collected evidence from various functions.
// Fields are always set once detection has run (even to empty values), so
// consumers can tell if detection ran and found nothing.
type Signals struct {
	// AgentEnv is a list of every env var match we found.
	// We do not read the variable's value.
	//
	// Sent verbatim as Attributes.AgentEnv.
	AgentEnv []string `json:"agent_env"`

	AgentAncestor *AncestorMatch `json:"agent_ancestor"`

	// Unattributed is every ancestor that neither table named.
	//
	// Local diagnostics only — ChainShape is the wire carrier for this information.
	Unattributed []UnattributedAncestor `json:"unattributed"`

	// IDEHost is the presence of a known IDE process in the terminal ancestry.
	// In testing we found that some IDEs with AI Chat interfaces may not show the agent
	// E.g. in VS Code the Copilot chat agent will spawn a new terminal session to run the command
	// This signal gives us some basis to determine the scale of under-reporting agent usage from IDEs.
	IDEHost *AncestorMatch `json:"ide_host"`

	// Wrappers names task runners and process wrappers sitting between the
	// agent and our cli. Entries are keys from a fixed table, not raw process names.
	Wrappers []WrapperMatch `json:"wrappers"`

	// ChainShape is the ancestry as one character per ancestor, nearest first:
	//
	//	a agent   e ide-host   i interpreter   s shell   t terminal
	//	w wrapper   r remote   n init   ? unknown
	//
	// E.g. "sasst" is shell → agent → shell → shell → terminal.
	//
	// This is the wire carrier for Unattributed and most of WalkMeta: ancestry
	// depth is its length, a completed walk ends in 'n' or 'r', a depth-capped
	// walk has length == defaultMaxDepth, and 'w' positions index Wrappers.
	ChainShape string `json:"chain_shape"`

	// Interactive is three separate bools here but one three-character string in
	// final payload sent; see Attributes.Interactive.
	Interactive Interactive `json:"interactive"`

	// CI holds normalized provider ids ("github-actions", "semaphore"), not the
	// variable names they were detected from.
	CI []string `json:"ci"`

	// CIGeneric means the bare CI=true var was set but no provider recognized from our list of ids.
	CIGeneric bool `json:"ci_generic"`
}

// WrapperMatch is a task runner or process wrapper in the ancestry.
type WrapperMatch struct {
	Name  string `json:"name"`
	Depth int    `json:"depth"`
}

// AncestorMatch is a vendor attribution for a proc. ancestor plus the evidence
//
// Fields are drawn from fixed tables — Name is only ever a key of
// procFingerprints, ArgvPattern only ever a pattern from cmdlineFingerprints —
// never free-form process text. Raw argv is never carried here.
type AncestorMatch struct {
	Vendor string `json:"vendor"`
	Depth  int    `json:"depth"`

	// Name is the fingerprint-table key the executable basename matched.
	Name string `json:"name,omitempty"`

	// ArgvPattern is the fingerprint-table pattern found in the argv.
	ArgvPattern string `json:"argv_pattern,omitempty"`
}

// MatchedOn is a derived summary of which evidence fired, for display and grouping
func (m AncestorMatch) MatchedOn() string {
	switch {
	case m.Name != "" && m.ArgvPattern != "":
		return "name+argv"
	case m.ArgvPattern != "":
		return "argv"
	default:
		return "name"
	}
}

// UnattributedAncestor records an ancestor that could have been hosting an
// agent and that neither table named.
//
// ArgvReadable separates two cases:
//   - readable, no match → our argv fingerprint table doesn't include this ancestor;
//     maybe it's a new or changed agent and a table update could fix it.
//   - not readable → permission denied reading argv for this ancestor
//     (cross-user, elevated, or protected) — no table update can fix this.
type UnattributedAncestor struct {
	Kind         procKind `json:"kind"`
	Depth        int      `json:"depth"`
	ArgvReadable bool     `json:"argv_readable"`
}

type Interactive struct {
	StdinTTY  bool `json:"stdin_tty"`
	StdoutTTY bool `json:"stdout_tty"`
	StderrTTY bool `json:"stderr_tty"`
}

// WalkMeta tells how far the walk got and why it stopped.
type WalkMeta struct {
	DepthReached int    `json:"depth_reached"`
	Truncated    bool   `json:"truncated"`
	StoppedAt    string `json:"stopped_at"`
	// CmdlineReads counts ancestors where argv was readable.
	CmdlineReads int `json:"cmdline_reads"`

	// LookupFailed records whether a proc lookup failed.
	// The walk breaks on the first failed lookup; StoppedAt ("lookup_error") says where.
	LookupFailed bool `json:"lookup_failed"`

	// Chain is raw observation — process names, pids and (optionally) redacted
	// argv for local debugging ONLY. Populated only when Options.KeepChain is set,
	// which no production path does.
	Chain []ChainEntry `json:"-"`
}

// ChainEntry is one ancestor as observed, for local diagnostics. Intentionally
// untagged: see WalkMeta.Chain. Name is still restricted to exact fingerprint-table
// keys so that even here we are not accumulating arbitrary process names.
type ChainEntry struct {
	Depth       int
	Pid         int
	Ppid        int
	Name        string
	Kind        procKind
	Vendor      string
	ArgvPattern string
	Unmatched   bool     // candidate host we could not attribute
	Cmdline     []string // redacted; only set when Options.ShowCmdlines is set
	Error       string
}

// ---------------------------------------------------------------------------
// Detection
// ---------------------------------------------------------------------------

type Options struct {
	MaxDepth     int
	Budget       time.Duration
	KeepChain    bool
	ShowCmdlines bool // include (redacted) argv in the chain; local diagnostics only
	Source       ProcSource
	Getenv       func(string) (string, bool)
	StartPid     int

	// IsTerminal is here for mock-ability so we can run synthetic-tree tests. It
	// takes a file descriptor and reports whether it is a terminal.
	IsTerminal func(fd uintptr) bool

	// Tables names which fingerprint table revision to fetch & use.
	// Empty means use the compiled-in tables (will set "builtin" flag below).
	Tables string
}

// builtinTables is the revision reported when not using the remote feature flag tables
const builtinTables = "builtin"

func Detect(opts Options) Result {
	start := time.Now()
	// Set defaults based on options passed. Intent is to be convenient for overriding in synthetic tests.
	if opts.MaxDepth == 0 {
		opts.MaxDepth = defaultMaxDepth
	}
	if opts.Budget == 0 {
		opts.Budget = defaultBudget
	}
	if opts.Source == nil {
		opts.Source = newProcSource()
	}
	if opts.Getenv == nil {
		opts.Getenv = os.LookupEnv
	}
	if opts.IsTerminal == nil {
		// Reuses the CLI's existing isatty wrapper, already mocked in pkg/mock.
		fs := &cliio.RealFileSystem{}
		opts.IsTerminal = fs.IsTerminal
	}
	// Parent PID
	if opts.StartPid == 0 {
		opts.StartPid = os.Getppid()
	}

	if opts.Tables == "" {
		opts.Tables = builtinTables
	}

	res := Result{Schema: "agentdetect/v1", Tables: opts.Tables}
	res.Signals.AgentEnv = detectEnv(opts.Getenv)
	res.Signals.CI, res.Signals.CIGeneric = detectCI(opts.Getenv)
	res.Signals.Interactive = detectTTY(opts.IsTerminal)

	w := walk(opts)
	res.Signals.AgentAncestor = w.agent
	res.Signals.IDEHost = w.ide
	res.Signals.Unattributed = w.unattributed
	res.Signals.Wrappers = w.wrappers
	res.Signals.ChainShape = w.shape
	res.Walk = w.meta

	res.TimingUs = time.Since(start).Microseconds()
	return res
}

// detectEnv returns the fingerprint keys that fired, in table order. See
// Signals.AgentEnv.
func detectEnv(getenv func(string) (string, bool)) []string {
	matches := make([]string, 0, 2)
	for _, v := range envFingerprints {
		// Presence-with-nonempty-value. A var set to "" or "0" is treated as
		// absent since some tools unset by emptying rather than deleting.
		// The value itself is never recorded anywhere.
		if val, ok := getenv(v); ok && truthy(val) {
			matches = append(matches, v)
		}
	}
	return matches
}

// detectCI returns normalized provider ids, plus whether a bare CI variable was
// set with no provider identified.
func detectCI(getenv func(string) (string, bool)) ([]string, bool) {
	providers := make([]string, 0, 1)
	for _, fp := range ciFingerprints {
		if v, ok := getenv(fp.Var); ok && truthy(v) {
			providers = append(providers, fp.Provider)
		}
	}
	generic := false
	if v, ok := getenv(genericCIVar); ok && truthy(v) {
		generic = len(providers) == 0
	}
	return providers, generic
}

// truthy is the shared reading of a boolean-ish environment variable.
// Also considered "unset" if empty or contains 0/false
func truthy(v string) bool {
	return v != "" && v != "0" && !strings.EqualFold(v, "false")
}

func detectTTY(isTerminal func(fd uintptr) bool) Interactive {
	return Interactive{
		StdinTTY:  isTerminal(os.Stdin.Fd()),
		StdoutTTY: isTerminal(os.Stdout.Fd()),
		StderrTTY: isTerminal(os.Stderr.Fd()),
	}
}

// walkResult groups the walk's outputs
type walkResult struct {
	agent        *AncestorMatch
	ide          *AncestorMatch
	unattributed []UnattributedAncestor
	wrappers     []WrapperMatch
	shape        string
	meta         WalkMeta
}

// walk climbs the ancestor chain, reporting the CLOSEST match
func walk(opts Options) walkResult {
	var (
		agent *AncestorMatch
		ide   *AncestorMatch
		// Empty rather than nil to distinguish "none" from "absent" for these.
		unattributed = make([]UnattributedAncestor, 0, 2)
		wrappers     = make([]WrapperMatch, 0, 2)
		shape        []byte
		meta         WalkMeta
		deadline     = time.Now().Add(opts.Budget)
		seen         = make([]int, 0, 8)
		pid          = opts.StartPid
		// childStart is the start time of the nearest descendant we could read
		childStart int64
	)

	for depth := 1; ; depth++ {
		if pid <= 1 {
			meta.StoppedAt = "init"
			break
		}
		if slices.Contains(seen, pid) {
			meta.StoppedAt = "cycle"
			break
		}
		if depth > opts.MaxDepth {
			meta.StoppedAt = "max_depth"
			meta.Truncated = true
			break
		}
		if time.Now().After(deadline) {
			meta.StoppedAt = "budget_exceeded"
			meta.Truncated = true
			break
		}
		seen = append(seen, pid)

		info, err := opts.Source.Info(pid)
		if err != nil {
			meta.LookupFailed = true
			if opts.KeepChain {
				meta.Chain = append(meta.Chain, ChainEntry{Depth: depth, Pid: pid, Kind: kindUnknown, Error: "lookup_failed"})
			}
			meta.StoppedAt = "lookup_error"
			break
		}

		// PID-reuse guard: a parent cannot have started after its child. If it
		// appears to have, the pid was recycled and this "ancestor" is unrelated.
		if childStart != 0 && info.StartTime != 0 && info.StartTime > childStart {
			meta.StoppedAt = "pid_reuse"
			meta.Truncated = true
			break
		}
		// Only carry forward a readable stamp; a 0 must not overwrite the last good one.
		if info.StartTime != 0 {
			childStart = info.StartTime
		}

		meta.DepthReached = depth
		if len(info.Cmdline) > 0 {
			meta.CmdlineReads++
		}

		// Name evidence: emittableName is the matched table KEY, not the raw basename.
		fp, emittableName, known := resolveFingerprint(info)
		if !known {
			fp.Kind = kindUnknown
		}

		entry := ChainEntry{
			Depth:  depth,
			Pid:    info.Pid,
			Ppid:   info.Ppid,
			Name:   emittableName,
			Kind:   fp.Kind,
			Vendor: fp.Vendor,
		}

		if fp.Kind == kindWrapper && emittableName != "" {
			wrappers = append(wrappers, WrapperMatch{Name: emittableName, Depth: depth})
		}

		nameVendor, nameKey := "", ""
		if fp.Kind == kindAgent {
			nameVendor, nameKey = fp.Vendor, emittableName
		}

		// Record the nearest editor in the ancestry.
		if fp.Kind == kindIDEHost && ide == nil {
			ide = &AncestorMatch{Vendor: fp.Vendor, Depth: depth, Name: emittableName}
		}

		// Argv evidence, for kinds whose argv states their own identity. Both tables
		// are consulted independently, so an agent named by basename and argv reports both.
		pattern, argvVendor := "", ""
		if argvEligible(fp.Kind) {
			pattern, argvVendor = matchCmdline(fp.Kind, info.Cmdline)
			entry.ArgvPattern = pattern
		}

		// Attribution: name wins the vendor on disagreement, but both evidence
		// fields survive so the disagreement stays visible downstream.
		if nameVendor != "" || argvVendor != "" {
			vendor := nameVendor
			if vendor == "" {
				vendor = argvVendor
			}
			entry.Kind = kindAgent
			entry.Vendor = vendor
			if agent == nil {
				agent = &AncestorMatch{
					Vendor:      vendor,
					Depth:       depth,
					Name:        nameKey,
					ArgvPattern: pattern,
				}
			}
		} else if argvEligible(fp.Kind) {
			// A candidate host we failed to name.
			entry.Unmatched = true
			unattributed = append(unattributed, UnattributedAncestor{
				Kind:         entry.Kind,
				Depth:        depth,
				ArgvReadable: len(info.Cmdline) > 0,
			})
		}

		// Shape records the final kind, after any argv reclassification.
		shape = append(shape, entry.Kind.code())

		if opts.KeepChain {
			if opts.ShowCmdlines {
				entry.Cmdline = redact(info.Cmdline)
			}
			meta.Chain = append(meta.Chain, entry)
		}

		// sshd is a hard boundary: ancestry above it is a different session.
		if fp.Kind == kindRemote {
			meta.StoppedAt = "remote_boundary"
			break
		}

		pid = info.Ppid
	}

	return walkResult{
		agent:        agent,
		ide:          ide,
		unattributed: unattributed,
		wrappers:     wrappers,
		shape:        string(shape),
		meta:         meta,
	}
}

// argvEligible reports whether an ancestor's argv states its own identity. It's
// an allow-list: a new procKind defaults to not being read.
func argvEligible(k procKind) bool {
	switch k {
	case kindAgent, kindInterpreter, kindUnknown:
		return true
	}
	return false
}

// matchCmdline returns the matched pattern and its vendor, both empty on a miss.
// Matching is confined to identity argv POSITIONS, a precision guard atop argvEligible.
func matchCmdline(kind procKind, argv []string) (string, string) {
	fields := argvIdentityFields(kind, argv)
	if len(fields) == 0 {
		return "", ""
	}
	// Normalize separators so Windows backslash paths match the forward-slash patterns.
	joined := strings.ToLower(strings.Join(fields, " "))
	joined = strings.ReplaceAll(joined, `\`, "/")
	for _, fp := range cmdlineFingerprints {
		if strings.Contains(joined, fp.Pattern) {
			return fp.Pattern, fp.Vendor
		}
	}
	return "", ""
}

// maxIdentityArgs is how many non-flag args after argv[0] can still be naming the
// program rather than describing its work: 2 for `node --flag /path/cli.js`, 3 for
// launcher forms like `uv tool run aider`. Past that, args are input, not identity.
const maxIdentityArgs = 3

// argvIdentityFields selects the argv positions worth matching for a kind.
// kindUnknown gets argv[0] only — its args are user data (grep, docker, a user
// script). Other eligible kinds are hosts whose first few args name what they run.
func argvIdentityFields(kind procKind, argv []string) []string {
	if len(argv) == 0 {
		return nil
	}
	if kind == kindUnknown {
		return argv[:1]
	}

	fields := []string{argv[0]}
	for _, arg := range argv[1:] {
		if len(fields) > maxIdentityArgs {
			break
		}
		// Skip rather than stop, so `node --flag /path/cli.js` stays matchable.
		if strings.HasPrefix(arg, "-") {
			continue
		}
		fields = append(fields, arg)
	}
	return fields
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func normalizeName(name string) string {
	name = strings.TrimSpace(name)
	// Split on both separators: a Windows path can reach a Unix build via a mock,
	// fixture, or WSL, and filepath.Base only knows the local separator.
	if i := strings.LastIndexAny(name, `/\`); i >= 0 {
		name = name[i+1:]
	}
	name = strings.ToLower(name)
	return strings.TrimSuffix(name, ".exe")
}

// redact scrubs known argv secrets before displayed locally or sent in telemetry.
var secretFlags = []string{
	"--password", "--secret", "--token", "--api-key", "--api-secret",
	"--client-secret", "--private-key", "--sasl-password",
	// -p over-redacts (mkdir -p, docker -p) — deliberate: err toward dropping, never leaking.
	"-p",
}

var secretKeySubstrings = []string{"password", "secret", "token", "apikey", "api_key", "credential", "passwd", "auth"}

func redact(argv []string) []string {
	out := make([]string, 0, len(argv))
	redactNext := false

	for _, arg := range argv {
		if redactNext {
			out = append(out, "<redacted>")
			redactNext = false
			continue
		}

		if key, _, found := strings.Cut(arg, "="); found {
			lowerKey := strings.ToLower(key)
			if slices.ContainsFunc(secretKeySubstrings, func(sub string) bool {
				return strings.Contains(lowerKey, sub)
			}) {
				out = append(out, key+"=<redacted>")
				continue
			}
		}

		if slices.Contains(secretFlags, strings.ToLower(arg)) {
			redactNext = true
		}

		out = append(out, arg)
	}
	return out
}
