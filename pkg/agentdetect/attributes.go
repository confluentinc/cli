package agentdetect

// ---------------------------------------------------------------------------
// Wire format
// ---------------------------------------------------------------------------

// Attributes is the payload sent to the usage service.
// See also Result, which carries more info for local diagnostics, debug chain
//
// A follow-up PR will assign these fields onto CliV1Usage, after the update to that repo.
// This also keeps the detection-to-wire mapping testable on its own.
type Attributes struct {
	// AgentEnvVars carries fingerprint KEYS (env var names), not vendor ids. See
	// Signals.AgentEnv.
	AgentEnvVars []string `json:"agent_env_vars,omitempty"`

	// AgentProc and AgentArgv are the evidence for the nearest agent ancestor:
	// the procFingerprints key its basename matched, and the cmdlineFingerprints
	// pattern found in its argv.
	AgentProc *string `json:"agent_proc,omitempty"`
	AgentArgv *string `json:"agent_argv,omitempty"`

	// IDEHost is the procFingerprints key of the editor in the ancestry; see Signals.IDEHost.
	IDEHost *string `json:"ide_host,omitempty"`

	// Interactive is a fixed three-character string, stdin/stdout/stderr:
	// "ioe" when all three are text terminals, "-" in place of any that is
	// not. A fully redirected invocation is "---".
	Interactive string `json:"interactive,omitempty"`

	// ChainShape encodes ancestry depth, every wrapper's position, and the
	// entire Unattributed population — kind from the character, depth from
	// the position. See Signals.ChainShape and Signals.Unattributed.
	ChainShape string `json:"chain_shape,omitempty"`

	// CmdWrappers are procFingerprints keys, nearest first. Depths aren't sent
	// separately: the 'w' positions in ChainShape correspond to this list in
	// order.
	CmdWrappers []string `json:"cmd_wrappers,omitempty"`

	// CiProviders holds normalized CI provider ids. A bare CI variable with no
	// recognizable provider appears as ciUnknown.
	CiProviders []string `json:"ci_providers,omitempty"`

	// Tables identifies the fingerprint table revision that produced everything
	// above. Since tables ship independently of CLI releases
	// it allows consumers to know if null AgentProc means "no agent" or
	// "not in the table yet".
	Tables string `json:"agent_tables,omitempty"`
}

// ciUnknown is the wire spelling of Signals.CIGeneric.
const ciUnknown = "unknown"

// Interactive attribute three-char string encoding.
const (
	ttyStdin  = 'i'
	ttyStdout = 'o'
	ttyStderr = 'e'
	ttyAbsent = '-'
)

// Attributes projects the Result onto the fields actually sent over the wire.
func (r Result) Attributes() Attributes {
	s := r.Signals

	attrs := Attributes{
		AgentEnvVars: nonEmpty(s.AgentEnv),
		Interactive:  encodeInteractive(s.Interactive),
		ChainShape:   s.ChainShape,
		CiProviders:  nonEmpty(s.CI),
		Tables:       r.Tables,
	}

	if a := s.AgentAncestor; a != nil {
		attrs.AgentProc = optional(string(a.Name))
		attrs.AgentArgv = optional(string(a.ArgvPattern))
	}
	if h := s.IDEHost; h != nil {
		attrs.IDEHost = optional(string(h.Name))
	}

	for _, w := range s.Wrappers {
		attrs.CmdWrappers = append(attrs.CmdWrappers, string(w.Name))
	}

	if s.CIGeneric {
		attrs.CiProviders = append(attrs.CiProviders, ciUnknown)
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
// stay distinguishable.
func optional(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// nonEmpty drops an empty slice (exists because we use it as a Signal during detection),
// so omitempty can leave out the field and reduce noise in the payload.
func nonEmpty(s []string) []string {
	if len(s) == 0 {
		return nil
	}
	return s
}
