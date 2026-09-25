package flink

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWithThousands(t *testing.T) {
	tests := map[int]string{
		0:        "0",
		5:        "5",
		999:      "999",
		1000:     "1,000",
		12345:    "12,345",
		1234567:  "1,234,567",
		10000000: "10,000,000",
		-1234:    "-1,234",
	}
	for in, want := range tests {
		require.Equal(t, want, withThousands(in))
	}
}

func TestQueryProgressPaintsAndClears(t *testing.T) {
	var buf bytes.Buffer
	p := newQueryProgressTo(&buf, true)
	p.throttle = 0 // don't rate-limit in the test

	p.start()
	p.update(1000)
	p.update(2500000)

	out := buf.String()
	require.Contains(t, out, "Running query...")
	require.Contains(t, out, "Fetched 1,000 rows...")
	require.Contains(t, out, "Fetched 2,500,000 rows...")
	require.Contains(t, out, "\r") // repaints in place

	before := buf.Len()
	p.clear()
	require.Greater(t, buf.Len(), before) // clear wrote the blanking sequence
	// After clearing, the tail is spaces + CR, leaving the line blank.
	require.True(t, strings.HasSuffix(buf.String(), "\r"))
}

func TestQueryProgressSilentWhenDisabled(t *testing.T) {
	var buf bytes.Buffer
	p := newQueryProgressTo(&buf, false)
	p.throttle = 0

	p.start()
	p.update(1000)
	p.clear()

	require.Empty(t, buf.String())
}

func TestQueryProgressThrottles(t *testing.T) {
	var buf bytes.Buffer
	p := newQueryProgressTo(&buf, true) // default throttle keeps rapid updates out

	p.start()
	p.update(1) // within throttle of start; suppressed
	p.update(2) // still within throttle; suppressed

	require.NotContains(t, buf.String(), "Fetched")
}
