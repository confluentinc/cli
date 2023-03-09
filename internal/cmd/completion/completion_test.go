package completion

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestCompletion(t *testing.T) {
	t.Parallel()

	for _, shell := range []string{"bash", "zsh"} {
		cmd := new(cobra.Command)
		out, err := completion(cmd, shell)
		require.NoError(t, err)
		require.Contains(t, out, fmt.Sprintf("# %s completion", shell))
	}
}
