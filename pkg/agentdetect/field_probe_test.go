package agentdetect

import (
	"encoding/json"
	"os"
	"runtime"
	"testing"
)

// probeVar gates the field probe. Everything else in this package runs against a
// synthetic tree and is hermetic; this one reads the REAL process ancestry and
// the REAL environment, so its output depends entirely on where it was run.
//
// Off by default rather than skipped-in-CI, because "CI" is not the condition —
// a developer running `make test` inside an agent would otherwise get a
// different result than the same command in a terminal, and a test whose output
// depends on who invoked it is not a test.
const probeVar = "AGENTDETECT_PROBE"

// TestFieldProbe is the Phase 6 instrument: it prints what detection sees on the
// machine, agent and OS it is run from. Phase 1 ships as dead code with no call
// site, so this is the only way to exercise the package against a real tree
// before the wiring exists.
//
//	AGENTDETECT_PROBE=1 go test ./pkg/agentdetect/ -run TestFieldProbe -v
//
// For a host with no Go toolchain — which is most Windows boxes and every
// borrowed machine — compile it and carry the binary over:
//
//	GOOS=windows GOARCH=amd64 go test -c ./pkg/agentdetect/ -o agentdetect-probe.exe
//	GOOS=linux   GOARCH=amd64 go test -c ./pkg/agentdetect/ -o agentdetect-probe
//	AGENTDETECT_PROBE=1 ./agentdetect-probe -test.run TestFieldProbe -test.v
//
// It asserts nothing. There is no correct answer to assert against: the point is
// to capture what a surface actually looks like, and a probe that failed on an
// unexpected tree would be discarding exactly the finding worth having. Read the
// output, and if it is a surface we have not seen, turn it into a fixture in
// ide_surfaces_test.go.
func TestFieldProbe(t *testing.T) {
	if os.Getenv(probeVar) == "" {
		t.Skipf("set %s=1 to probe the real process tree", probeVar)
	}

	// KeepChain and ShowCmdlines are the local-diagnostics options — the reason
	// they exist. The chain is what explains a miss: an ancestor typed `unknown`
	// with readable argv means a missing table row, and one with no argv at all
	// means a permission boundary no table can fix. Argv is redacted here and
	// WalkMeta.Chain is unserializable, so neither can reach a payload.
	res := Detect(Options{KeepChain: true, ShowCmdlines: true})

	t.Logf("platform: %s/%s", runtime.GOOS, runtime.GOARCH)
	t.Logf("elapsed: %dus", res.TimingUs)
	t.Logf("tables: %s", res.Tables)

	// The raw observation, re-read here rather than taken from the chain.
	//
	// ChainEntry.Name is the table KEY that matched, so an unattributed ancestor
	// reports "" — correct for the chain, useless for diagnosing a miss, which is
	// the entire job of this probe. Without the observed basename beside it there
	// is no way to tell a name we have no row for from a name we failed to
	// normalize. Local only, and it goes no further than this log line.
	src := newProcSource()
	observed := func(pid int) string {
		info, err := src.Info(pid)
		if err != nil {
			return "<unreadable>"
		}
		return info.Name
	}

	t.Log("--- chain (local only, never transmitted) ---")
	for _, e := range res.Walk.Chain {
		raw := observed(e.Pid)
		t.Logf("  depth=%d pid=%d kind=%-13s key=%-22q vendor=%-14q argv_pattern=%q",
			e.Depth, e.Pid, e.Kind, e.Name, e.Vendor, e.ArgvPattern)
		t.Logf("        observed: %q -> normalized %q", raw, normalizeName(raw))
		if e.Unmatched {
			t.Logf("        MISS: argv-eligible, nothing attributed")
		}
		if len(e.Cmdline) > 0 {
			t.Logf("        argv[0]: %q", e.Cmdline[0])
		}
		if e.Error != "" {
			t.Logf("        error: %s", e.Error)
		}
	}

	t.Log("--- walk health ---")
	t.Logf("  depth_reached=%d stopped_at=%s truncated=%v cmdline_reads=%d errors=%d",
		res.Walk.DepthReached, res.Walk.StoppedAt, res.Walk.Truncated,
		res.Walk.CmdlineReads, res.Walk.Errors)

	t.Log("--- signals (diagnostic shape) ---")
	t.Log("\n" + indent(t, res.Signals))

	// The payload is printed separately and last, because it is the thing under
	// review. Everything above is context for reading it.
	t.Log("--- attributes (what would actually be sent) ---")
	t.Log("\n" + indent(t, res.Attributes()))
}

func indent(t *testing.T, v any) string {
	t.Helper()
	blob, err := json.MarshalIndent(v, "  ", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return "  " + string(blob)
}
