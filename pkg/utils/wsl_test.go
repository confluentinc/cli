package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsWSLKernelRelease(t *testing.T) {
	tests := []struct {
		name     string
		release  string
		expected bool
	}{
		{
			name:     "WSL2 kernel release",
			release:  "5.15.90.1-microsoft-standard-WSL2\n",
			expected: true,
		},
		{
			name:     "WSL1 kernel release",
			release:  "4.4.0-19041-Microsoft\n",
			expected: true,
		},
		{
			name:     "native Linux kernel release",
			release:  "6.5.0-25-generic\n",
			expected: false,
		},
		{
			name:     "empty release",
			release:  "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, isWSLKernelRelease(tt.release))
		})
	}
}
