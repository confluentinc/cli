package agentdetect

// ProcInfo is the minimum a process source must supply for detection.
//
// Cmdline is best-effort on every platform: it is always permission-gated, never
// guaranteed so an empty Cmdline is normal, not an error.
type ProcInfo struct {
	Pid  int
	Ppid int

	// Name is the executable name as the platform reports it — a full exec path,
	// a bare basename, any casing, with or without a .exe suffix. The walk runs
	// normalizeName (idempotent) before every lookup, so a source that fails
	// to normalize just produces a clean miss rather than a wrong answer.
	Name string

	Cmdline []string

	// StartTime is an opaque, monotonically-increasing process creation stamp,
	// used only to catch pid reuse (a parent that started after its child isn't
	// really its parent). Zero disables the check rather than failing it.
	StartTime int64
}

// ProcSource abstracts process lookup so the walk can be unit-tested against a
// synthetic tree — the real implementation is a thin shim, tests inject a fake.
type ProcSource interface {
	Info(pid int) (ProcInfo, error)
}
