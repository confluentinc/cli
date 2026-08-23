package agentdetect

import (
	"encoding/json"
	"os"
	"runtime"
	"testing"
)

// This file is a manual field probe, not part of the automated test suite.
// TestFieldProbe runs detection against the REAL process ancestry and
// environment of the machine it is launched on and prints what it found; it
// asserts nothing. Use it to see what an agent or IDE surface actually looks
// like in practice, and to turn any process tree we have not seen into a fixture
// in ide_surfaces_test.go.

// probeVar gates the probe. The output depends entirely on where it was run, so
// it is set as opt-in rather than skipped-in-CI.
const probeVar = "AGENTDETECT_PROBE"

// TestFieldProbe prints what detection sees on the machine, agent and OS it is
// run from. To run it, set the probeVar and run the test:
//
//	AGENTDETECT_PROBE=1 go test ./pkg/agentdetect/ -run TestFieldProbe -v
//
// For a host with no Go toolchain — most Windows boxes and every borrowed
// machine — compile it and carry the binary over:
//
//	GOOS=windows GOARCH=amd64 go test -c ./pkg/agentdetect/ -o agentdetect-probe.exe
//	GOOS=linux   GOARCH=amd64 go test -c ./pkg/agentdetect/ -o agentdetect-probe
//	AGENTDETECT_PROBE=1 ./agentdetect-probe -test.run TestFieldProbe -test.v
func TestFieldProbe(t *testing.T) {
	if os.Getenv(probeVar) == "" {
		t.Skipf("set %s=1 to probe the real process tree", probeVar)
	}

	// KeepChain and ShowCmdlines are the local-diagnostics options.
	// Argv is redacted here and WalkMeta.Chain is unserializable, so neither can reach a payload.
	res := Detect(Options{KeepChain: true, ShowCmdlines: true})

	t.Logf("platform: %s/%s", runtime.GOOS, runtime.GOARCH)
	t.Logf("elapsed: %dus", res.TimingUs)
	t.Logf("tables: %s", res.Tables)

	// The raw observation of name, re-read here rather than taken from the chain.
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
	t.Logf("  depth_reached=%d stopped_at=%s truncated=%v cmdline_reads=%d lookup_failed=%v",
		res.Walk.DepthReached, res.Walk.StoppedAt, res.Walk.Truncated,
		res.Walk.CmdlineReads, res.Walk.LookupFailed)

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
