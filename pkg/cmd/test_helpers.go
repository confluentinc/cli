package cmd

import (
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/featureflags"
)

// ExecuteCommand runs the root command with the given args, and returns the Cobra command output and the standard output as strings, or returns an error.
func ExecuteCommand(root *cobra.Command, args ...string) (string, error) {
	if args == nil {
		args = []string{}
	}

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	os.Stdout = w
	root.SetOut(w)
	defer func() {
		os.Stdout = oldStdout
		root.SetOut(oldStdout)
	}()

	root.SetArgs(args)
	cfg := &config.Config{IsTest: true, Contexts: map[string]*config.Context{}}
	featureflags.Init(cfg)
	if _, err := root.ExecuteC(); err != nil {
		return "", err
	}

	w.Close()
	output, err := io.ReadAll(r)

	return string(output), err
}
