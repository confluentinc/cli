package log

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestLogger_SetLoggingVerbosity_Shorthand(t *testing.T) {
	type fields struct {
		Logger  *Logger
		Command string
	}
	tests := []struct {
		name   string
		cmd    string
		want   Level
	}{
		{
			name: "default logging level",
			cmd: "ccloud command-with-dash -b --other",
			want: ERROR,
		},
		{
			name: "warn logging level",
			cmd: "ccloud command-with-dash -b -v --other",
			want: WARN,
		},
		{
			name: "info logging level",
			cmd: "ccloud command-with-dash -b -vv --other",
			want: INFO,
		},
		{
			name: "debug logging level",
			cmd: "ccloud command-with-dash -b -vvv --other",
			want: DEBUG,
		},
		{
			name: "trace logging level",
			cmd: "ccloud command-with-dash -b -vvvv --other",
			want: TRACE,
		},
		{
			name: "trace logging level more than four flags",
			cmd: "ccloud command-with-dash -b -vvvvvvvvvvvvvvvvv --other",
			want: TRACE,
		},
		{
			name: "other flag in between",
			cmd: "ccloud command-with-dash -vvbv",
			want: DEBUG,
		},
		{
			name: "other flag before and after",
			cmd: "ccloud command-with-dash -bvvvb",
			want: DEBUG,
		},
		{
			name: "multiple groups",
			cmd: "ccloud command-with-dash -v -v -v",
			want: DEBUG,
		},
		{
			name: "invalid",
			cmd: "ccloud command-with-dash -vv-vv",
			want: ERROR,
		},
		{
			name: "invalid 2",
			cmd: "ccloud command-with-dash -vvvv-",
			want: ERROR,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New()
			args := strings.Split(tt.cmd, " ")
			SetLoggingVerbosity(args, logger)
			require.Equal(t, tt.want, logger.GetLevel())
		})
	}
}

func TestLogger_SetLoggingVerbosity_FullFlag(t *testing.T) {
	type fields struct {
		Logger  *Logger
		Command string
	}
	tests := []struct {
		name   string
		cmd    string
		want   Level
	}{
		{
			name: "default logging level",
			cmd: "ccloud command-with-dash -b --other",
			want: ERROR,
		},
		{
			name: "warn logging level",
			cmd: "ccloud command-with-dash -b --verbose --other",
			want: WARN,
		},
		{
			name: "info logging level",
			cmd: "ccloud command-with-dash -b --verbose --verbose --other",
			want: INFO,
		},
		{
			name: "debug logging level",
			cmd: "ccloud command-with-dash -b --verbose --verbose --verbose --other",
			want: DEBUG,
		},
		{
			name: "trace logging level",
			cmd: "ccloud command-with-dash -b --verbose --verbose --verbose --verbose --verbose --other",
			want: TRACE,
		},
		{
			name: "trace logging level more than four flags",
			cmd: "ccloud command-with-dash -b --verbose --verbose --verbose --verbose --verbose --verbose --other",
			want: TRACE,
		},
		{
			name: "invalid",
			cmd: "ccloud command-with-dash --verbosenot",
			want: ERROR,
		},
		{
			name: "with shorthand flag",
			cmd: "ccloud command-with-dash --verbose -vv",
			want: DEBUG,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New()
			args := strings.Split(tt.cmd, " ")
			SetLoggingVerbosity(args, logger)
			require.Equal(t, tt.want, logger.GetLevel())
		})
	}
}
