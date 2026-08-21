package agentdetect

// ---------------------------------------------------------------------------
// Wire format
// ---------------------------------------------------------------------------

// Attributes is the payload sent to the usage service. Result carries more
// (local diagnostics, vendor ids, a debug chain) but only fields defined here
// are ever transmitted — keeping the two types separate makes that a checkable
// property rather than a comment.
//
// It intentionally does not import the usage SDK; a later phase assigns these
// fields onto CliV1Usage separately, so the detection-to-wire mapping is
// testable on its own.
//
// Pointers where absence is a finding, values where the field is always
// known. Every field is optional, so an older service ignores what it does
// not recognize and an older CLI simply omits everything.
type Attributes struct {
	// AgentEnv carries fingerprint KEYS (env var names, or "file:<path>" for
	// filesystem markers), not vendor ids. See Signals.AgentEnv.
	AgentEnv []string `json:"agent_env,omitempty"`

	// AgentProc and AgentArgv are the evidence for the nearest agent ancestor:
	// the procFingerprints key its basename matched, and the cmdlineFingerprints
	// pattern found in its argv. Either may be absent; at least one is present
	// whenever an agent was attributed.
	//
	// Keys, not vendor ids: unlike AgentEnv's markers, a basename or argv
	// pattern doesn't always spell out its vendor ("q" is Amazon Q), so a
	// vendor mapping does exist (vendorForProcKey, vendorForArgvPattern) —
	// but the fingerprint tables move independently of CLI releases, so a
	// baked-in vendor id would go stale, while a key can be re-resolved (and
	// corrected) at query time. The cost is that the headline agent-share
	// number needs a join against Tables; the benefit is that a mis-attributed
	// key is fixable in the warehouse instead of permanent.
	AgentProc *string `json:"agent_proc,omitempty"`
	AgentArgv *string `json:"agent_argv,omitempty"`

	// IDEHost is the procFingerprints key of the editor in the ancestry — "code",
	// "idea" — not the vendor id it maps to. Means "agent-capable environment",
	// never "agent-initiated"; see Signals.IDEHost.
	IDEHost *string `json:"ide_host,omitempty"`

	// Interactive is a fixed three-character string, stdin/stdout/stderr in that
	// order: "ioe" when all three are terminals, "-" in place of any that is
	// not. A fully redirected invocation is "---".
	//
	// Kept as three positions rather than collapsed into one bool because the
	// disagreements are informative — "i--" is a human piping output to a file,
	// "-oe" reads more like a harness — and an AND across the three would erase
	// that. stderr is included because stdout-redirected-but-stderr-a-terminal
	// is the ordinary shape of a human capturing output, and isn't
	// distinguishable from CI without it.
	Interactive string `json:"interactive,omitempty"`

	// ChainShape encodes ancestry depth, whether the walk completed (ends in
	// 'n' or 'r'), every wrapper's position, and the entire Unattributed
	// population — kind from the character, depth from the position. See
	// Signals.ChainShape and Signals.Unattributed.
	ChainShape string `json:"chain_shape,omitempty"`

	// Wrappers are procFingerprints keys, nearest first. Depths aren't sent
	// separately: the 'w' positions in ChainShape correspond to this list in
	// order.
	Wrappers []string `json:"wrappers,omitempty"`

	// CI holds normalized provider ids. A bare CI variable with no recognizable
	// provider appears as ciUnknown.
	//
	// Signals keeps that case in its own bool (CIGeneric) because CI co-occurs
	// with every specific provider and would otherwise poison a set meant to
	// distinguish providers. Safe to fold in here because Detect only sets it
	// when no provider was named, so it never appears alongside a real one.
	CI []string `json:"ci,omitempty"`

	// Tables identifies the fingerprint table revision that produced everything
	// above (builtinTables until tables ship via feature flag independently of
	// CLI releases — at which point a null AgentProc could mean "no agent" or
	// "not in the table yet", and this field is what lets a consumer tell the
	// two apart).
	Tables string `json:"agent_tables,omitempty"`
}

// ciUnknown is the wire spelling of Signals.CIGeneric.
const ciUnknown = "unknown"

// Interactive encoding. A fixed-width string, so position is meaningful and an
// absent stream is a placeholder rather than a shorter string.
const (
	ttyStdin  = 'i'
	ttyStdout = 'o'
	ttyStderr = 'e'
	ttyAbsent = '-'
)

// Attributes projects a Result onto the fields actually sent over the wire.
// It reads only Signals, never a Vendor field — those exist on Result for
// local debugging, and reading one here would reintroduce the frozen-
// attribution problem keys exist to avoid. TestAttributesEmitKeysNotVendors
// pins this.
func (r Result) Attributes() Attributes {
	s := r.Signals

	attrs := Attributes{
		AgentEnv:    nonEmpty(s.AgentEnv),
		Interactive: encodeInteractive(s.Interactive),
		ChainShape:  s.ChainShape,
		CI:          nonEmpty(s.CI),
		Tables:      r.Tables,
	}

	if a := s.AgentAncestor; a != nil {
		attrs.AgentProc = optional(a.Name)
		attrs.AgentArgv = optional(a.ArgvPattern)
	}
	if h := s.IDEHost; h != nil {
		attrs.IDEHost = optional(h.Name)
	}

	for _, w := range s.Wrappers {
		attrs.Wrappers = append(attrs.Wrappers, w.Name)
	}

	if s.CIGeneric {
		attrs.CI = append(attrs.CI, ciUnknown)
	}

	return attrs
}

// encodeInteractive renders the three TTY bits as a fixed-width string.
func encodeInteractive(i Interactive) string {
	flag := func(on bool, c byte) byte {
		if on {
			return c
		}
		return ttyAbsent
	}
	return string([]byte{
		flag(i.StdinTTY, ttyStdin),
		flag(i.StdoutTTY, ttyStdout),
		flag(i.StderrTTY, ttyStderr),
	})
}

// optional returns nil for the empty string, so "absent" and "present but empty"
// stay distinguishable on the wire.
func optional(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// nonEmpty drops an empty slice so omitempty can elide the field. Detection
// guarantees these are non-nil, which is right for Signals — a count of zero is
// a finding there — but on the wire an empty list is a column of empty lists.
func nonEmpty(s []string) []string {
	if len(s) == 0 {
		return nil
	}
	return s
}
