package kafka

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParsePrincipal(t *testing.T) {
	id, err := parsePrincipal("User:u-12345")
	require.NoError(t, err)
	require.Equal(t, "u-12345", id)
}

func TestParsePrincipal_NoPrefix(t *testing.T) {
	_, err := parsePrincipal("u-12345")
	require.Error(t, err)
}

func TestParsePrincipal_NumericId(t *testing.T) {
	_, err := parsePrincipal("User:12345")
	require.Error(t, err)
}
