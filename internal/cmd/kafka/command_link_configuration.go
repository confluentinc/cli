package kafka

import (
	"github.com/spf13/cobra"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

func (c *linkCommand) newConfigurationCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "configuration",
		Short: "Manage inter-cluster link configurations.",
	}

	if cfg.IsCloudLogin() {
		cmd.AddCommand(c.newConfigurationDescribeCommand())
		cmd.AddCommand(c.newConfigurationUpdateCommand())
	} else {
		cmd.AddCommand(c.newConfigurationDescribeCommandOnPrem())
		cmd.AddCommand(c.newConfigurationUpdateCommandOnPrem())
	}

	return cmd
}
