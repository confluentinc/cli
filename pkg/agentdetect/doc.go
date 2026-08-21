// Package agentdetect reports evidence that the CLI was invoked by an AI coding
// agent rather than typed by a human.
//
// It collects two independent signals, chosen because their failure modes
// barely overlap:
//
//   - Environment markers the agent vendors leave: variables they set
//     (CLAUDECODE, CURSOR_AGENT, …). Reported as fingerprint-table keys, never
//     vendor ids — see Signals.AgentEnv.
//   - The process ancestor tree, matched by executable basename and by the
//     identity positions of each ancestor's argv.
//
// An editor in the ancestry (Signals.IDEHost) is reported as "agent-capable
// environment", not "agent-initiated" — a human typing in the integrated
// terminal is indistinguishable from an agent tool call, so the two are not
// separated here. In-editor agents that set no env var and run in-process
// (JetBrains, Zed) or as a bare node child are the main recall gap; the env
// signal is what covers the known ones.
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
