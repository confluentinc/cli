// Package agentdetect reports evidence that the CLI was invoked by an AI coding
// agent rather than typed by a human.
//
// It collects two independent signals, chosen because their failure modes
// barely overlap:
//
//   - Environment markers the agent vendors leave: variables they set
//     (CLAUDECODE, CURSOR_AGENT, …) and filesystem markers for the ones that
//     identify by path instead. Reported as fingerprint-table keys, never
//     vendor ids — see Signals.AgentEnv.
//   - The process ancestor tree, matched by executable basename and by the
//     identity positions of each ancestor's argv.
//
// Provenance (e.g. "an editor's extension host spawned us") is reported
// separately from vendor attribution, since it's often the only claim
// available and is a weaker one. Its strength is platform-dependent and must
// not be pooled in analysis: macOS gives each editor child role its own
// executable name, so agent-initiated and human-typed calls inside an editor
// are cleanly separable; Linux and Windows re-exec one binary for every role
// with identical argv, collapsing the two into a single reported value
// (Signals.IDESpawn, kindIDENodeHost). Non-Chromium editors (JetBrains, Zed)
// run the agent in-process and offer no separation on any platform.
//
// Detect returns a Result, where every field is either a value from this
// package's own fixed vocabulary (a vendor id, a procKind, a fingerprint-table
// key) or a count — never an observed process name, path, or argument. The one
// exception, WalkMeta.Chain, is local-debugging-only and unserializable.
//
// Result is not the wire format: Result.Attributes projects it onto
// Attributes, which is, so adding a diagnostic to Result cannot add a column
// to the payload. Attributes carries fingerprint-table KEYS rather than
// resolved vendor ids, since the tables ship independently of CLI releases and
// a baked-in vendor id would go stale.
//
// Nothing here decides whether the caller "is an agent" — that inference
// belongs downstream in analytics, where it can be revised without a release.
//
// Detection is bounded and non-fatal by construction: a wall-clock budget, a
// depth cap, a cycle guard and a pid-reuse guard each stop the walk and report
// how far it got. It has no failure mode that can affect the command the user
// actually ran.
package agentdetect
