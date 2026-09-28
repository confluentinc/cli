package schemaregistry

import (
	"github.com/spf13/cobra"
)

// No RunRequirement annotation: like its sibling command_exporter_configuration_describe.go,
// this connects directly to a Schema Registry endpoint via --schema-registry-endpoint and
// certificate flags, with no MDS login required.
func (c *command) newExporterConfigurationClusterLinkCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster-link",
		Short: "Manage the schema exporter's Cluster Link config.",
	}

	cmd.AddCommand(c.newExporterConfigurationClusterLinkDescribeCommandOnPrem())

	return cmd
}
