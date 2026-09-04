// Package agentdetect reports evidence that the CLI was invoked by an AI coding
// agent rather than typed by a human.
//
// It collects two independent signals, chosen because their failure modes
// barely overlap:
//
//   - Environment markers the agent vendors leave: variables they set
//     (CLAUDECODE, CURSOR_AGENT, …) — see Signals.AgentEnv.
//   - The process ancestor tree, agent vendor matches by executable basename
//     and/or the identity positions of each ancestor's argv.
//
// A known code editor in the ancestry (Signals.IDEHost) is reported as an
// "agent-capable environment", but not necessarily "agent-initiated" —
// a human typing in the integrated terminal is indistinguishable from an agent
// in many tested cases, so the two are not separated here.
// In-editor agents that set no env var and run in-process (JetBrains, Zed)
// or as a bare node child are the main recall gap; the env signal helps cover the known ones.
//
// Detect returns a Result, where every field is either a value from this
// package's own fixed list (a vendor id, a procKind, a fingerprint-table key)
// or a count — never a raw observed process name, path, or argument. The one
// exception, WalkMeta.Chain, is local-debugging-only and unserializable.
//
// Result is not the wire format: Result.Attributes projects it onto
// Attributes for attaching to usage events, so adding a diagnostic to Result cannot add a column
// to the payload. Attributes carries fingerprint-table KEYS; the tables will
// ship independently of CLI releases and can be updated without a CLI release.
//
// Nothing in this package makes a claim about whether the caller "is an agent" —
// that inference is left to downstream analytics to resolve.
//
// Detection is bounded and non-fatal by construction: a wall-clock budget, a
// depth cap, a cycle guard and a pid-reuse guard each stop the walk and report
// how far it got. It has no failure mode that can affect the command the user
// actually ran.
package agentdetect
