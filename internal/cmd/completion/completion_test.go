package completion

import (
	"testing"

	"github.com/stretchr/testify/require"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

func TestCompletionBash(t *testing.T) {
	req := require.New(t)

	root := pcmd.BuildRootCommand()
	cmd := NewCompletionCmd(root,"ccloud")
	root.AddCommand(cmd)

	output, err := pcmd.ExecuteCommand(root, "completion", "bash")
	req.NoError(err)
	req.Contains(output, "bash completion for")
}

func TestCompletionZsh(t *testing.T) {
	req := require.New(t)

	root := pcmd.BuildRootCommand()
	cmd := NewCompletionCmd(root, "ccloud")
	root.AddCommand(cmd)

	output, err := pcmd.ExecuteCommand(root, "completion", "zsh")
	req.Error(err)
	req.Contains(output, "Error: unsupported shell type \"zsh\"")
}

func TestCompletionUnknown(t *testing.T) {
	req := require.New(t)

	root := pcmd.BuildRootCommand()
	cmd := NewCompletionCmd(root, "ccloud")
	root.AddCommand(cmd)

	output, err := pcmd.ExecuteCommand(root, "completion", "newsh")
	req.Error(err)
	req.Contains(output, "Error: unsupported shell type \"newsh\"")
}

func TestCompletionNone(t *testing.T) {
	req := require.New(t)

	root := pcmd.BuildRootCommand()
	cmd := NewCompletionCmd(root, "ccloud")
	root.AddCommand(cmd)

	output, err := pcmd.ExecuteCommand(root, "completion")
	req.Error(err)
	req.Contains(output, "Error: accepts 1 arg(s), received 0")
}
