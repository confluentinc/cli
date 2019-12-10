package config

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
)

type command struct {
	*pcmd.CLICommand
	prerunner pcmd.PreRunner
}

// New returns the Cobra command for `config`.
func New(config *config.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cliCmd := pcmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "config",
			Short: "Modify the CLI config files.",
		},
		config, prerunner)
	cmd := &command{CLICommand: cliCmd, prerunner: prerunner}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	c.AddCommand(NewContext(c.Config, c.prerunner))
}
