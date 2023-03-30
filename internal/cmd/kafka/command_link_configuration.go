package kafka

import (
	"github.com/spf13/cobra"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

func (c *linkCommand) newConfigurationCommand(cfg *v1.Config) *cobra.Command {
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
