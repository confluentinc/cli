package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newExporterConfigurationClusterLinkDescribeCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe the schema exporter's Cluster Link config.",
		Long:  "Derive the Cluster Link config(s) that replicate a schema exporter's subject/context translation, so they don't have to be hand-copied.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.exporterConfigurationClusterLinkDescribe,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	addCaLocationAndClientPathFlags(cmd)
	addSchemaRegistryEndpointFlag(cmd)
	pcmd.AddOutputFlagWithDefaultValue(cmd, output.JSON.String())

	return cmd
}
