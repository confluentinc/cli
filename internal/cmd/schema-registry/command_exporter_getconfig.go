package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newExporterGetConfigCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-config <name>",
		Short: "Get the schema exporter configuration.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.exporterGetConfig,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		addCaLocationFlag(cmd)
		addSchemaRegistryEndpointFlag(cmd)
	}
	pcmd.AddOutputFlagWithDefaultValue(cmd, output.JSON.String())

	if cfg.IsCloudLogin() {
		// Deprecated
		pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
		cobra.CheckErr(cmd.Flags().MarkHidden("api-key"))

		// Deprecated
		pcmd.AddApiSecretFlag(cmd)
		cobra.CheckErr(cmd.Flags().MarkHidden("api-secret"))
	}

	return cmd
}

func (c *command) exporterGetConfig(cmd *cobra.Command, args []string) error {
	client, err := c.GetSchemaRegistryClient(cmd)
	if err != nil {
		return err
	}

	configs, err := client.GetExporterConfig(args[0])
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(configs)
	return table.Print()
}
