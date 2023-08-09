package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
)

func (c *command) newClusterCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "cluster",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLoginOrOnPremLogin},
	}

	if cfg.IsCloudLogin() {
		cmd.Short = "Manage Schema Registry cluster."
		cmd.Long = "Manage the Schema Registry cluster for the current environment."
	} else {
		cmd.Short = "Manage Schema Registry clusters."
	}

	cmd.AddCommand(c.newClusterDeleteCommand())
	cmd.AddCommand(c.newClusterDescribeCommand())
	cmd.AddCommand(c.newClusterEnableCommand())
	cmd.AddCommand(c.newClusterListCommandOnPrem())
	cmd.AddCommand(c.newClusterUpdateCommand())
	cmd.AddCommand(c.newClusterUpgradeCommand())

	return cmd
}
