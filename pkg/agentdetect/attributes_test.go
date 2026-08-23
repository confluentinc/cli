package agentdetect

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"
)

// The wire format is where a privacy or interpretability mistake becomes
// permanent, so these tests assert what LEAVES the process rather than what
// detection computed.

// The whole point of emitting keys: a payload must carry the evidence, and the
// vendor must be recoverable from it rather than baked in. Vendors deliberately
// disagree with their keys here so that reading the wrong field cannot pass.
func TestAttributesEmitKeysNotVendors(t *testing.T) {
	res := Result{
		Tables: builtinTables,
		Signals: Signals{
			AgentAncestor: &AncestorMatch{
				Vendor:      "claude-code",
				Name:        "claude",
				ArgvPattern: "@anthropic-ai/claude-code",
			},
			IDEHost: &AncestorMatch{Vendor: "vscode", Name: "code"},
		},
	}

	attrs := res.Attributes()

	if attrs.AgentProc == nil || *attrs.AgentProc != "claude" {
		t.Errorf("agent_proc = %v, want the table key %q", attrs.AgentProc, "claude")
	}
	if attrs.AgentArgv == nil || *attrs.AgentArgv != "@anthropic-ai/claude-code" {
		t.Errorf("agent_argv = %v, want the table pattern", attrs.AgentArgv)
	}
	// "code" is the key, "vscode" is the vendor.
	if attrs.IDEHost == nil || *attrs.IDEHost != "code" {
		t.Errorf("ide_host = %v, want the table key %q not the vendor", attrs.IDEHost, "code")
	}

	// And the vendors must not have leaked in anywhere else.
	blob, err := json.Marshal(attrs)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(blob), "vscode") {
		t.Errorf("payload contains the vendor id %q: %s", "vscode", blob)
	}
}

// A key that resolves to nothing is a payload nobody can interpret. This is the
// standing check that the emittable key spaces and the vendor tables have not
// drifted apart.
func TestEveryEmittableAncestryKeyResolves(t *testing.T) {
	for key, fp := range procFingerprints {
		vendor, known := vendorForProcKey(key)
		if !known {
			t.Errorf("procFingerprints key %q does not resolve", key)
			continue
		}
		// Agents and editors are the kinds whose keys reach the wire as
		// agent_proc and ide_host, and an attribution with no vendor behind it
		// is not an attribution.
		if (fp.Kind == kindAgent || fp.Kind == kindIDEHost) && vendor == "" {
			t.Errorf("key %q is kind %s but resolves to no vendor", key, fp.Kind)
		}
	}

	for _, fp := range cmdlineFingerprints {
		vendor, known := vendorForArgvPattern(fp.Pattern)
		if !known || vendor == "" {
			t.Errorf("argv pattern %q resolves to (%q, %v), want a vendor", fp.Pattern, vendor, known)
		}
	}
}

// The Unattributed bound is computable in production only because ChainShape
// carries it. That rests on two table invariants nothing else enforces, and
// breaking either would leave the documented derivation silently wrong.
func TestChainShapeCarriesTheUnattributedPopulation(t *testing.T) {
	// Invariant 1: no kindAgent row without a vendor. One would attribute
	// nothing, land in Unattributed, and record 'a' — a character the
	// derivation does not look for, so the entry would be uncountable.
	for key, fp := range procFingerprints {
		if fp.Kind == kindAgent && fp.Vendor == "" {
			t.Errorf("agent row %q has no vendor: it would land in Unattributed as 'a'", key)
		}
	}

	// Invariant 2: the shape characters for the argv-eligible kinds are distinct
	// from each other and from every other kind's, or a character in the shape
	// would not identify which population an entry came from.
	eligible := map[byte]procKind{}
	all := []procKind{
		kindAgent, kindIDEHost,
		kindInterpreter, kindShell, kindTerminal, kindWrapper, kindRemote, kindInit, kindUnknown,
	}
	for _, k := range all {
		if !argvEligible(k) || k == kindAgent {
			continue
		}
		if prev, dup := eligible[k.code()]; dup {
			t.Errorf("kinds %s and %s share shape character %q", prev, k, k.code())
		}
		eligible[k.code()] = k
	}
	for _, k := range all {
		if argvEligible(k) || k == kindAgent {
			continue
		}
		if other, clash := eligible[k.code()]; clash {
			t.Errorf("ineligible kind %s shares character %q with eligible %s", k, k.code(), other)
		}
	}

	// And the derivation itself, end to end: node is an interpreter whose argv
	// names nothing, so it is unattributed; the shell and terminal are not.
	src, start := tree("zsh", "node", "ghostty")
	res := detect(t, src, start, nil)

	if got := res.Signals.ChainShape; got != "sit" {
		t.Fatalf("chain_shape = %q, want %q", got, "sit")
	}

	var recovered []UnattributedAncestor
	for i := 0; i < len(res.Signals.ChainShape); i++ {
		if k, ok := eligible[res.Signals.ChainShape[i]]; ok {
			recovered = append(recovered, UnattributedAncestor{Kind: k, Depth: i + 1})
		}
	}
	// Guard against the whole comparison passing on two empty lists.
	if len(res.Signals.Unattributed) == 0 {
		t.Fatal("scenario produced no unattributed ancestors; the derivation below would be vacuous")
	}
	if len(recovered) != len(res.Signals.Unattributed) {
		t.Fatalf("recovered %d entries from the shape, Unattributed has %d",
			len(recovered), len(res.Signals.Unattributed))
	}
	for i, want := range res.Signals.Unattributed {
		if recovered[i].Kind != want.Kind || recovered[i].Depth != want.Depth {
			t.Errorf("entry %d: recovered %+v from the shape, want %+v", i, recovered[i], want)
		}
	}
}

// Fixed width, and position is meaningful. The disagreements are the reason this
// is not one boolean, so each is spelled out.
func TestInteractiveEncoding(t *testing.T) {
	tests := []struct {
		name        string
		in          Interactive
		want        string
		description string
	}{
		{"all", Interactive{true, true, true}, "ioe", "human at a terminal"},
		{"none", Interactive{false, false, false}, "---", "CI or a fully redirected harness"},
		{"stdout redirected", Interactive{true, false, true}, "i-e", "human capturing output to a file"},
		{"stdin piped", Interactive{false, true, true}, "-oe", "a harness feeding the CLI"},
	}
	for _, test := range tests {
		if got := encodeInteractive(test.in); got != test.want {
			t.Errorf("%s (%s): encodeInteractive = %q, want %q", test.name, test.description, got, test.want)
		}
	}

	// Always three characters, so a consumer can index rather than parse.
	if got := len(encodeInteractive(Interactive{})); got != 3 {
		t.Errorf("encoding is %d characters, want a fixed 3", got)
	}
}

// CIGeneric is its own bool in Signals; on the wire it folds in.
func TestGenericCIFoldsIntoTheProviderList(t *testing.T) {
	src, start := tree("bash")

	res := detect(t, src, start, map[string]string{"CI": "true"})
	if got := res.Attributes().CI; !slices.Equal(got, []string{ciUnknown}) {
		t.Errorf("bare CI: ci = %v, want [%s]", got, ciUnknown)
	}

	res = detect(t, src, start, map[string]string{"CI": "true", "GITHUB_ACTIONS": "true"})
	attrs := res.Attributes()
	if !slices.Equal(attrs.CI, []string{"github-actions"}) {
		t.Errorf("named provider: ci = %v, want [github-actions]", attrs.CI)
	}
	if slices.Contains(attrs.CI, ciUnknown) {
		t.Errorf("ci = %v: %q must never appear beside a named provider", attrs.CI, ciUnknown)
	}
}

// The boundary this type exists to enforce. Local diagnostics are what a
// hand-rolled mapping would have shipped by accident.
func TestAttributesCarryNoLocalDiagnostics(t *testing.T) {
	src, start := tree("zsh", "claude", "ghostty")
	res := detect(t, src, start, map[string]string{"CLAUDECODE": "1"})

	blob, err := json.Marshal(res.Attributes())
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	if err := json.Unmarshal(blob, &got); err != nil {
		t.Fatal(err)
	}

	allowed := []string{
		"agent_env", "agent_proc", "agent_argv", "ide_host",
		"interactive", "chain_shape", "wrappers", "ci", "agent_tables",
	}
	for key := range got {
		if !slices.Contains(allowed, key) {
			t.Errorf("payload carries unapproved attribute %q — every field is a schema decision", key)
		}
	}

	// Named individually because each is a specific thing that must not ship:
	// walk internals, timings, the debug chain, and the vendor conveniences.
	for _, banned := range []string{"walk", "timing_us", "schema", "signals", "unattributed", "vendor", "depth"} {
		if strings.Contains(string(blob), `"`+banned+`"`) {
			t.Errorf("payload contains %q: %s", banned, blob)
		}
	}
}

// A null agent_proc means "no agent" or "not in the tables yet", and only this
// field separates them once the tables ship out of band.
func TestTablesRevisionIsAlwaysReported(t *testing.T) {
	src, start := tree("zsh")

	res := detect(t, src, start, nil)
	if got := res.Attributes().Tables; got != builtinTables {
		t.Errorf("agent_tables = %q, want %q when no revision was supplied", got, builtinTables)
	}

	res = Detect(Options{
		Source: src, StartPid: start, Getenv: env(nil),
		IsTerminal: notATTY, Tables: "remote-rev-1",
	})
	if got := res.Attributes().Tables; got != "remote-rev-1" {
		t.Errorf("agent_tables = %q, want the supplied revision", got)
	}
}

// Absence is a finding for the ancestry fields, so they must be null rather than
// empty — and an argv-only attribution must not manufacture a proc key.
func TestAbsentEvidenceIsNullNotEmpty(t *testing.T) {
	attrs := Result{Signals: Signals{
		AgentAncestor: &AncestorMatch{Vendor: "codex", ArgvPattern: "@openai/codex"},
	}}.Attributes()

	if attrs.AgentProc != nil {
		t.Errorf("agent_proc = %q, want null: the basename matched nothing", *attrs.AgentProc)
	}
	if attrs.AgentArgv == nil {
		t.Fatal("agent_argv = null, want the pattern")
	}
	if attrs.IDEHost != nil {
		t.Error("ide_host should be null when no editor was found")
	}
}
