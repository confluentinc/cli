package cmd

import (
	"bytes"

	"github.com/spf13/cobra"
)

// ExecuteCommand runs the root command with the given args, and returns the output string or an error.
func ExecuteCommand(root *cobra.Command, args ...string) (output string, err error) {
	if args == nil {
		args = []string{}
	}
	_, output, err = ExecuteCommandC(root, args...)
	return output, err
}

// ExecuteCommandC runs the root command with the given args, and returns the executed command and the output string or an error.
func ExecuteCommandC(root *cobra.Command, args ...string) (c *cobra.Command, output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs(args)

	c, err = root.ExecuteC()

	return c, buf.String(), err
}

// BuildRootCommand creates a new root command for testing
func BuildRootCommand() *cobra.Command {
	return &cobra.Command{}
}
