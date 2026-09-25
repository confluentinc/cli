package flink

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/term"
)

// progressThrottle bounds how often the row count is repainted, so a fast drain
// doesn't spend its time writing to the terminal.
const progressThrottle = 150 * time.Millisecond

// queryProgress renders a single, self-overwriting status line while a query
// drains, so the command isn't silent for the minutes a large result can take.
//
// It writes only to its writer (stderr in production) and only when that writer
// is a terminal, so piped or redirected output — and stdout — stay untouched. It
// repaints in place with a carriage return and blanks the previous line with
// spaces rather than ANSI erase codes, which keeps it portable across terminals.
type queryProgress struct {
	w        io.Writer
	enabled  bool
	throttle time.Duration
	now      func() time.Time

	last    time.Time
	lastLen int
	painted bool
}

// newQueryProgress builds a progress line bound to stderr, enabled only when
// stderr is an interactive terminal.
func newQueryProgress() *queryProgress {
	return newQueryProgressTo(os.Stderr, term.IsTerminal(int(os.Stderr.Fd())))
}

// newQueryProgressTo is the injectable constructor used by tests.
func newQueryProgressTo(w io.Writer, enabled bool) *queryProgress {
	return &queryProgress{
		w:        w,
		enabled:  enabled,
		throttle: progressThrottle,
		now:      time.Now,
	}
}

// start paints the initial status, covering the wait for the statement to leave
// PENDING and the first fetch, before any rows exist to count.
func (p *queryProgress) start() {
	if !p.enabled {
		return
	}
	p.paint("Running query...")
	p.last = p.now()
}

// update repaints the running row count. It satisfies query.Options.OnProgress
// and is a no-op between throttle intervals.
func (p *queryProgress) update(rows int) {
	if !p.enabled {
		return
	}
	if p.painted && p.now().Sub(p.last) < p.throttle {
		return
	}
	p.last = p.now()
	p.paint(fmt.Sprintf("Fetched %s rows...", withThousands(rows)))
}

// clear erases the status line so the result prints on a clean line. Callers must
// call it before writing anything else (the result, a warning, an error).
func (p *queryProgress) clear() {
	if !p.enabled || !p.painted {
		return
	}
	fmt.Fprintf(p.w, "\r%s\r", strings.Repeat(" ", p.lastLen))
	p.lastLen = 0
	p.painted = false
}

// paint overwrites the current line with s, blanking any trailing characters left
// over from a longer previous line.
func (p *queryProgress) paint(s string) {
	pad := ""
	if p.lastLen > len(s) {
		pad = strings.Repeat(" ", p.lastLen-len(s))
	}
	fmt.Fprintf(p.w, "\r%s%s", s, pad)
	p.lastLen = len(s)
	p.painted = true
}

// withThousands formats n with comma thousands separators (1234567 -> "1,234,567").
func withThousands(n int) string {
	s := strconv.Itoa(n)
	sign := ""
	if strings.HasPrefix(s, "-") {
		sign, s = "-", s[1:]
	}

	var b strings.Builder
	for i, digit := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(digit)
	}
	return sign + b.String()
}
