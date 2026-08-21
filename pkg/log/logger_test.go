package log

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogger_Flush(t *testing.T) {
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

func TestLogger_FlushAfterRaisingVerbosity(t *testing.T) {
	for _, verbosity := range []Level{DEBUG, TRACE, UNSAFE_TRACE} {
		t.Run(fmt.Sprintf("raised to %d", verbosity), func(t *testing.T) {
			buf := new(bytes.Buffer)
			l := New(ERROR, buf)

			l.Debug("hi there")
			l.SetVerbosity(int(verbosity))
			l.Flush()

			require.Contains(t, buf.String(), "hi there")
		})
	}
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
		{name: "environment variable is clamped", env: "99", want: UNSAFE_TRACE},
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
