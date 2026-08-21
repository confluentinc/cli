package agentdetect

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// The rest of this package's tests drive Detect() against synthetic trees, which
// deliberately never touch the platform. These two exercise the real
// gopsutil-backed source against the live process tree instead.
//
// They exist because the proof of concept's Windows source compiled for three
// architectures and had never once been executed, and "it builds" was mistaken
// for "it works" as a result. A source that returns an error for every ancestor
// still produces a confident, well-formed, empty report — so the assertions here
// are about the walk reaching real processes, not about what it finds in them.
//
// Nothing here asserts a vendor. What the tree contains depends on who is running
// the tests and from where; asserting on it would make this fail in CI, or pass for
// the wrong reason on a developer's laptop.

func TestLiveSourceReadsTheRealProcessTree(t *testing.T) {
	info, err := newProcSource().Info(os.Getpid())
	if err != nil {
		t.Fatalf("reading our own process failed: %v", err)
	}

	if info.Pid != os.Getpid() {
		t.Errorf("Pid = %d, want %d", info.Pid, os.Getpid())
	}
	if info.Ppid != os.Getppid() {
		t.Errorf("Ppid = %d, want %d", info.Ppid, os.Getppid())
	}
	if info.Name == "" {
		t.Error("Name is empty — the exec-path read and the name fallback both failed")
	}

	// StartTime gates the pid-reuse guard. Zero disables it, which is exactly the
	// condition that left the guard silently off on Windows with the hand-rolled
	// source, so it is worth failing on rather than tolerating.
	if info.StartTime == 0 {
		t.Error("StartTime is 0 — the pid-reuse guard is disabled on this platform")
	}
}

func TestLiveWalkTerminatesAndReachesAnAncestor(t *testing.T) {
	res := Detect(Options{KeepChain: true})

	if res.Walk.DepthReached < 1 {
		t.Fatalf("DepthReached = %d, want >= 1; stopped_at = %q, lookup_failed = %v",
			res.Walk.DepthReached, res.Walk.StoppedAt, res.Walk.LookupFailed)
	}

	// A real walk from a test binary terminates at init, at a terminal, or at the
	// remote boundary. Landing on lookup_error at depth 1 is the signature of a
	// source that cannot read anything — the failure this test is here to catch.
	if res.Walk.StoppedAt == "lookup_error" && res.Walk.DepthReached <= 1 {
		t.Errorf("walk failed immediately: stopped_at = %q at depth %d", res.Walk.StoppedAt, res.Walk.DepthReached)
	}

	if len(res.Signals.ChainShape) != res.Walk.DepthReached {
		t.Errorf("ChainShape %q has %d entries, DepthReached = %d",
			res.Signals.ChainShape, len(res.Signals.ChainShape), res.Walk.DepthReached)
	}

	// Guard health is always populated, whatever the walk found.
	if res.Walk.StoppedAt == "" {
		t.Error("StoppedAt is empty; every exit path must record why the walk stopped")
	}
	if res.TimingUs <= 0 {
		t.Errorf("TimingUs = %d, want a positive measurement", res.TimingUs)
	}
}

// The real source, driven through a real walk, must not put an observed process
// name or argument into anything serializable — the same invariant the synthetic
// test pins, checked once against live data where the names are real.
func TestLiveResultSerializesNoRawObservations(t *testing.T) {
	res := Detect(Options{KeepChain: true, ShowCmdlines: true})
	if len(res.Walk.Chain) == 0 {
		t.Skip("walk produced no chain entries on this platform")
	}

	encoded, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	payload := string(encoded)

	for _, e := range res.Walk.Chain {
		// Names that ARE fingerprint-table keys are vocabulary and may legitimately
		// appear; anything else must not.
		if e.Name == "" {
			continue
		}
		if _, ok := procFingerprints[e.Name]; ok {
			continue
		}
		if strings.Contains(payload, e.Name) {
			t.Errorf("serialized Result leaks observed process name %q", e.Name)
		}
	}
	if strings.Contains(payload, `"chain":`) {
		t.Errorf("serialized Result contains the diagnostic chain:\n%s", payload)
	}
}
