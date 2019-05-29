package config

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/config"
)

type command struct {
	*cobra.Command
	config *config.Config
}

// New returns the Cobra command for `config`.
func New(config *config.Config) *cobra.Command {
	cmd := &command{
		Command: &cobra.Command{
			Use:   "config",
			Short: "Modify the CLI config files.",
		},
		config: config,
	}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	c.AddCommand(NewContext(c.config))
}
