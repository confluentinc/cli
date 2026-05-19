package debug

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

type command struct {
	*pcmd.CLICommand
}

func New(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "debug",
		Short:  "Debug utilities for internal testing.",
		Hidden: true,
	}

	c := &command{CLICommand: pcmd.NewAnonymousCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newPanicCommand())

	return cmd
}
