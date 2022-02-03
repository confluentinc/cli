package context

import (
	"testing"

	"github.com/stretchr/testify/require"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

func TestNewRow_Current(t *testing.T) {
	ctx := &v1.Context{
		Name:           "context",
		PlatformName:   "platform",
		CredentialName: "credential",
	}

	expected := &row{
		Current:    "*",
		Name:       "context",
		Platform:   "platform",
		Credential: "credential",
	}

	require.Equal(t, expected, newRow(true, ctx, "context"))
}

func TestNewRow_NotCurrent(t *testing.T) {
	ctx := &v1.Context{
		Name:           "context",
		PlatformName:   "platform",
		CredentialName: "credential",
	}

	expected := &row{
		Current:    "",
		Name:       "context",
		Platform:   "platform",
		Credential: "credential",
	}

	require.Equal(t, expected, newRow(true, ctx, "other"))
}

func TestFormatCurrent(t *testing.T) {
	require.Equal(t, "*", formatCurrent(true, true))
	require.Equal(t, "", formatCurrent(true, false))
	require.Equal(t, "true", formatCurrent(false, true))
	require.Equal(t, "false", formatCurrent(false, false))
}
