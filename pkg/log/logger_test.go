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

func TestLogger_SetVerbosity(t *testing.T) {
	tests := []struct {
		name        string
		verbosity   int
		env         string
		want        Level
		wantWarning bool
	}{
		{name: "no flag and no environment variable", want: ERROR},
		{name: "flag only", verbosity: int(DEBUG), want: DEBUG},
		{name: "environment variable only", env: "3", want: DEBUG},
		{name: "flag wins over environment variable", verbosity: int(WARN), env: "4", want: WARN},
		{name: "environment variable at the cap", env: "4", want: TRACE},
		{name: "unsafe-trace level is not reachable from the environment", env: "5", want: ERROR, wantWarning: true},
		{name: "out-of-range environment variable warns and is ignored", env: "99", want: ERROR, wantWarning: true},
		{name: "unparsable environment variable warns and is ignored", env: "debug", want: ERROR, wantWarning: true},
		{name: "negative environment variable warns and is ignored", env: "-1", want: ERROR, wantWarning: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv(VerbosityEnvVar, test.env)
			out := new(bytes.Buffer)
			l := New(ERROR, out)

			l.SetVerbosity(test.verbosity)

			require.Equal(t, test.want, l.Level)
			if test.wantWarning {
				require.Contains(t, out.String(), "[WARN]")
				require.Contains(t, out.String(), VerbosityEnvVar)
			} else {
				require.Empty(t, out.String())
			}
		})
	}
}
