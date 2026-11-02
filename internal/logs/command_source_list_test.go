package logs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormatScopes(t *testing.T) {
	require.Equal(t, "organization, environment, resource", formatScopes([]string{"ORGANIZATION", "ENVIRONMENT", "RESOURCE"}))
	require.Equal(t, "cloud-cluster", formatScopes([]string{"CLOUD_CLUSTER"}))
	require.Equal(t, "", formatScopes(nil))
}

func TestDashIfEmpty(t *testing.T) {
	require.Equal(t, "-", dashIfEmpty(""))
	require.Equal(t, "lkc", dashIfEmpty("lkc"))
}
