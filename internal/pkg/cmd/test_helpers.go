package cmd

import (
	"bytes"

	"github.com/spf13/cobra"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/featureflags"
)

// ExecuteCommand runs the root command with the given args, and returns the output string or an error.
func ExecuteCommand(root *cobra.Command, args ...string) (string, error) {
	if args == nil {
		args = []string{}
	}
	_, output, err := ExecuteCommandC(root, args...)
	return output, err
}

// ExecuteCommandC runs the root command with the given args, and returns the executed command and the output string or an error.
func ExecuteCommandC(root *cobra.Command, args ...string) (*cobra.Command, string, error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs(args)

	cfg := &v1.Config{IsTest: true, Contexts: map[string]*v1.Context{}}
	featureflags.Init(cfg)

	c, err := root.ExecuteC()
	return c, buf.String(), err
}

// BuildRootCommand creates a new root command for testing
func BuildRootCommand() *cobra.Command {
	return &cobra.Command{}
}
