package kafka

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/config"
)

func (c *linkCommand) newConfigurationCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "configuration",
		Short: "Manage cluster link configurations.",
	}

	if cfg.IsCloudLogin() {
		cmd.AddCommand(c.newConfigurationListCommand())
		cmd.AddCommand(c.newConfigurationUpdateCommand())
	} else {
		cmd.AddCommand(c.newConfigurationListCommandOnPrem())
		cmd.AddCommand(c.newConfigurationUpdateCommandOnPrem())
	}

	return cmd
}
