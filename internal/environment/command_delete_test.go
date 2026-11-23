package environment

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPromptLabels(t *testing.T) {
	names := map[string]string{"env-1": "prod", "env-2": ""}
	require.Equal(t, []string{"env-1 (prod)", "env-2", "env-3"}, promptLabels([]string{"env-1", "env-2", "env-3"}, names))
}
