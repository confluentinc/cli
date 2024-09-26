package kafka

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/config"
)

func (c *command) newConfigurationCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "configuration",
		Short: "Manage topic configuration.",
	}

	if cfg.IsCloudLogin() {
		cmd.AddCommand(c.newConfigurationListCommand())
	} else {
		cmd.AddCommand(c.newConfigurationListCommandOnPrem())
	}

	return cmd
}
