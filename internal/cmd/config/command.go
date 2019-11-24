package config

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
)

type command struct {
	*cobra.Command
	config    *config.Config
	prerunner pcmd.PreRunner
	analytics analytics.Client
}

// New returns the Cobra command for `config`.
func New(config *config.Config, prerunner pcmd.PreRunner, analytics analytics.Client) *cobra.Command {
	cmd := &command{
		Command: &cobra.Command{
			Use:   "config",
			Short: "Modify the CLI config files.",
		},
		config:    config,
		prerunner: prerunner,
		analytics: analytics,
	}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	c.AddCommand(NewContext(c.config, c.prerunner, c.analytics))
}
