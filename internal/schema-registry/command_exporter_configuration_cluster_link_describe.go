package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newExporterConfigurationClusterLinkDescribeCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe the schema exporter's Cluster Link config.",
		Long:  "Derive the Cluster Link config(s) that replicate a schema exporter's subject/context translation, so they don't have to be hand-copied.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.exporterConfigurationClusterLinkDescribe,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		addCaLocationAndClientPathFlags(cmd)
	}
	addSchemaRegistryEndpointFlag(cmd)
	pcmd.AddOutputFlagWithDefaultValue(cmd, output.JSON.String())

	return cmd
}

func (c *command) exporterConfigurationClusterLinkDescribe(cmd *cobra.Command, args []string) error {
	client, err := c.GetSchemaRegistryClient(cmd)
	if err != nil {
		return err
	}

	configs, err := client.GetExporterClusterLinkConfig(args[0])
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(configs)
	return table.Print()
}
