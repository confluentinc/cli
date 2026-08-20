package schemaregistry

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/config"
)

func (c *command) newExporterConfigurationClusterLinkCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster-link",
		Short: "Manage the schema exporter's Cluster Link config.",
	}

	cmd.AddCommand(c.newExporterConfigurationClusterLinkDescribeCommand(cfg))

	return cmd
}
