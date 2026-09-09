package log

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoggerFlush(t *testing.T) {
	tests := []struct {
		name     string
		level    Level
		wantEmit bool
	}{
		{
			name:     "emit message that should be emitted",
			level:    TRACE,
			wantEmit: true,
		},
		{
			name:     "buffer messages that shouldn't be emitted",
			level:    ERROR,
			wantEmit: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			l := New(test.level, buf)
			l.Debug("hi there")
			if test.wantEmit {
				require.Len(t, l.buffer, 0)
			} else {
				require.Len(t, l.buffer, 1)
			}
			l.Flush()
			require.Len(t, l.buffer, 0)
			if test.wantEmit {
				require.Contains(t, buf.String(), "hi there")
			} else {
				require.Empty(t, buf.String())
			}
		})
	}
}

func TestLoggerFlushEmitsAtOrBelowThreshold(t *testing.T) {
	// Logged at the least-verbose level, so every line buffers instead of emitting; the threshold
	// is then raised before the flush.
	buffered := map[Level]string{
		WARN:         "warn: rare but handled",
		INFO:         "info: steady-state",
		DEBUG:        "debug: low-level detail",
		TRACE:        "trace: action tracing",
		UNSAFE_TRACE: "unsafe: sensitive detail",
	}

	buf := new(bytes.Buffer)
	l := New(ERROR, buf)
	l.Warn(buffered[WARN])
	l.Info(buffered[INFO])
	l.Debug(buffered[DEBUG])
	l.Trace(buffered[TRACE])
	l.UnsafeTrace(buffered[UNSAFE_TRACE])
	require.Len(t, l.buffer, len(buffered))

	l.SetVerbosity(int(DEBUG))
	l.Flush()

	// A buffered line emits only when its level is at or below the threshold. DEBUG sits on the
	// boundary, so an off-by-one comparison would drop it.
	out := buf.String()
	for level, message := range buffered {
		if level <= DEBUG {
			require.Contains(t, out, message)
		} else {
			require.NotContains(t, out, message)
		}
	}
	require.Empty(t, l.buffer)
}
