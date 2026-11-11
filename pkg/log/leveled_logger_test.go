package log

import (
	"bytes"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLeveledLogger(t *testing.T) {
	const signed = "https://example.com/things?token=secret"
	urlErr := &url.Error{Op: "Get", URL: signed, Err: errors.New("connection reset by peer")}

	tests := []struct {
		name  string
		level Level
		log   func(LeveledLogger)
		want  string
	}{
		{
			name:  "url value",
			level: DEBUG,
			log:   func(l LeveledLogger) { l.Debug("performing request", "method", "GET", "url", signed) },
			want:  "[DEBUG] performing request: method=GET url=https://example.com/things\n",
		},
		{
			name:  "error at trace",
			level: TRACE,
			log:   func(l LeveledLogger) { l.Error("request failed", "error", urlErr) },
			want:  `[DEBUG] request failed: error="Get \"https://example.com/things\": connection reset by peer"` + "\n",
		},
		{
			name:  "retry description with non-string values",
			level: DEBUG,
			log: func(l LeveledLogger) {
				l.Debug("retrying request", "request", "GET "+signed+" (status: 503)", "timeout", 1100*time.Millisecond, "remaining", 3)
			},
			want: `[DEBUG] retrying request: request="GET https://example.com/things (status: 503)" timeout=1.1s remaining=3` + "\n",
		},
		{
			name:  "unsafe trace",
			level: UNSAFE_TRACE,
			log:   func(l LeveledLogger) { l.Warn("performing request", "url", signed) },
			want:  "[DEBUG] performing request: url=https://example.com/things?token=secret\n",
		},
		{
			name:  "odd count",
			level: DEBUG,
			log:   func(l LeveledLogger) { l.Info("performing request", "method") },
			want:  "[DEBUG] performing request: EXTRA_VALUE_AT_END=method\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			original := CliLogger
			t.Cleanup(func() { CliLogger = original })
			var buf bytes.Buffer
			CliLogger = New(ERROR, &buf)
			CliLogger.SetVerbosity(int(test.level))

			test.log(LeveledLogger{})
			require.Contains(t, buf.String(), test.want)
		})
	}
}

func TestLeveledLogger_DefaultLevel(t *testing.T) {
	original := CliLogger
	t.Cleanup(func() { CliLogger = original })
	var buf bytes.Buffer
	CliLogger = New(ERROR, &buf)

	LeveledLogger{}.Error("request failed", "error", "connection reset by peer")
	require.Empty(t, buf.String(), "a failed attempt printed at the default ERROR level")
}
