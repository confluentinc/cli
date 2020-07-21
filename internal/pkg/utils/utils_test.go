package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContains(t *testing.T) {
	req := require.New(t)
	req.True(Contains([]string{"a"}, "a"))
}

func TestDoesNotContain(t *testing.T) {
	req := require.New(t)
	req.False(Contains([]string{}, "a"))
}
