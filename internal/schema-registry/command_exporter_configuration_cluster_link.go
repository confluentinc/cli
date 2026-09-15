package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

func (c *command) newExporterConfigurationClusterLinkCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "cluster-link",
		Short:       "Manage the schema exporter's Cluster Link config.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	cmd.AddCommand(c.newExporterConfigurationClusterLinkDescribeCommand())

	return cmd
}
