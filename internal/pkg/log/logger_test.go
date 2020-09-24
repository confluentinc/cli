package log

import (
	"bytes"
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			l := NewWithParams(&Params{
				Level:  tt.level,
				Output: buf,
				JSON:   false,
			})
			l.Debug("hi there")
			if tt.wantEmit {
				require.Len(t, l.buffer, 0)
			} else {
				require.Len(t, l.buffer, 1)
			}
			l.Flush()
			require.Len(t, l.buffer, 0)
			if tt.wantEmit {
				require.Contains(t, buf.String(), "hi there")
			} else {
				require.Empty(t, buf.String())
			}
		})
	}
}
