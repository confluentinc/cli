package ksql

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
)

type command struct {
	*cobra.Command
	config      *config.Config
	prerunner   pcmd.PreRunner
}

// New returns the default command object for interacting with KSQL.
func New(prerunner pcmd.PreRunner, config *config.Config) *cobra.Command {
	cmd := &command{
		Command: &cobra.Command{
			Use:               "ksql",
			Short:             "Manage KSQL.",
			PersistentPreRunE: prerunner.Authenticated(config),
		},
		config:      config,
		prerunner:   prerunner,
	}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	c.AddCommand(NewClusterCommand(c.config, c.prerunner))
}
