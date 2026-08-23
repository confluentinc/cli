package agentdetect

import (
	"context"
	"time"

	"github.com/shirou/gopsutil/v4/process"
)

// gopsutilSource is the real ProcSource: one implementation for darwin, linux
// and windows. Backed by github.com/shirou/gopsutil/v4/process.
type gopsutilSource struct {
	// readTimeout bounds a single ancestor read, best-effort only: several
	// calls (ExeWithContext, CmdlineSliceWithContext, PpidWithContext) ignore
	// the context on darwin, since they're sysctl/proc_pidpath calls with no
	// cancellation point. The load-bearing bound is the budget check in walk();
	// this is an extra guard on top of it.
	readTimeout time.Duration
}

const defaultReadTimeout = 10 * time.Millisecond

func newProcSource() ProcSource {
	// Avoids re-reading /proc/stat per ancestor on Linux to get boot time;
	// no-op on darwin and windows.
	process.EnableBootTimeCache(true)
	return gopsutilSource{readTimeout: defaultReadTimeout}
}

func (s gopsutilSource) Info(pid int) (ProcInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.readTimeout)
	defer cancel()

	p, err := process.NewProcessWithContext(ctx, int32(pid))
	if err != nil {
		return ProcInfo{}, err
	}

	// Ppid empty makes the entry useless, so it fails.
	ppid, err := p.PpidWithContext(ctx)
	if err != nil {
		return ProcInfo{}, err
	}

	info := ProcInfo{Pid: pid, Ppid: int(ppid)}

	// Prefer the full executable path over the kernel-supplied name: some
	// installs (i.e. Claude Code) name the binary by version, so the identity
	// is in the directory not the basename; resolveFingerprint uses it by
	// reading argv[0] and then path segments.
	//
	// Handed over raw — the walk normalizes via normalizeName
	if exe, err := p.ExeWithContext(ctx); err == nil && exe != "" {
		info.Name = exe
	} else if name, err := p.NameWithContext(ctx); err == nil {
		info.Name = name
	}

	// Best-effort: permission-gated across a user/elevation boundary. An empty
	// Cmdline is normal, not an error; the walk records the gap via
	// WalkMeta.CmdlineReads.
	if argv, err := p.CmdlineSliceWithContext(ctx); err == nil {
		info.Cmdline = argv
	}

	// Milliseconds since the epoch, uniform across platforms; zero disables the
	// pid-reuse check for that step rather than failing it.
	if created, err := p.CreateTimeWithContext(ctx); err == nil {
		info.StartTime = created
	}

	return info, nil
}
