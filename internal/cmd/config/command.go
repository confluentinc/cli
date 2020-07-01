package config

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type command struct {
	*pcmd.CLICommand
	prerunner pcmd.PreRunner
	analytics analytics.Client
}

// New returns the Cobra command for `config`.
func New(prerunner pcmd.PreRunner, analytics analytics.Client) *cobra.Command {
	cliCmd := pcmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "config",
			Short: "Modify the CLI configuration.",
		}, prerunner)
	cmd := &command{
		CLICommand: cliCmd,
		prerunner:  prerunner,
		analytics:  analytics,
	}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	c.AddCommand(NewContext(c.prerunner, c.analytics))
}
