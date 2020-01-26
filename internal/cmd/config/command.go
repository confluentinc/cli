package config

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
)

type command struct {
	*pcmd.CLICommand
	prerunner pcmd.PreRunner
	analytics analytics.Client
	config    *config.Config
}

// New returns the Cobra command for `config`.
func New(config *config.Config, prerunner pcmd.PreRunner, analytics analytics.Client) *cobra.Command {
	cliCmd := pcmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "config",
			Short: "Modify the CLI config files.",
		},
		config, prerunner)
	cmd := &command{
		CLICommand: cliCmd,
		prerunner:  prerunner,
		analytics:  analytics,
		config:     config,
	}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	c.AddCommand(NewContext(c.config, c.prerunner, c.analytics))
}
