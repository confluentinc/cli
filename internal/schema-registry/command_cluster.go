package schemaregistry

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/config"
)

func (c *command) newClusterCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use: "cluster",
	}

	if cfg.IsCloudLogin() {
		cmd.Short = "Manage Schema Registry cluster."
		cmd.Long = "Manage the Schema Registry cluster for the current environment."
	} else {
		cmd.Short = "Manage Schema Registry clusters."
	}

	cmd.AddCommand(c.newClusterDescribeCommand())
	cmd.AddCommand(c.newClusterListCommandOnPrem())
	cmd.AddCommand(c.newClusterUpdateCommand())

	return cmd
}
