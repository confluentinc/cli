package config

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/shared"
)

type command struct {
	*cobra.Command
	config *shared.Config
}

// New returns the Cobra command for `config`.
func New(config *shared.Config) *cobra.Command {
	cmd := &command{
		Command: &cobra.Command{
			Use:   "config",
			Short: "Manage config.",
		},
		config: config,
	}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	// remove redundant help command
	c.SetHelpCommand(&cobra.Command{
		Use:    "no-help",
		Hidden: true,
	})

	c.AddCommand(NewContext(c.config))
}
