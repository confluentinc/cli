package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *exporterCommand) newGetConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-config <name>",
		Short: "Get the configurations of the schema exporter.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.getConfig),
	}

	pcmd.AddOutputFlagWithDefaultValue(cmd, output.JSON.String())

	return cmd
}

func (c *exporterCommand) getConfig(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}

	configs, _, err := srClient.DefaultApi.GetExporterConfig(ctx, args[0])
	if err != nil {
		return err
	}

	outputFormat, err := cmd.Flags().GetString("output")
	if err != nil {
		return err
	}

	return output.StructuredOutputForCommand(cmd, outputFormat, configs)
}
