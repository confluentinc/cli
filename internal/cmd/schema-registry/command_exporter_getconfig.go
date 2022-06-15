package schemaregistry

import (
	"context"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *exporterCommand) newGetConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-config <name>",
		Short: "Get the configurations of the schema exporter.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.getConfig,
	}

	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlagWithDefaultValue(cmd, output.JSON.String())

	return cmd
}

func (c *exporterCommand) getConfig(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := getApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}

	return getExporterConfig(cmd, args[0], srClient, ctx)
}

func getExporterConfig(cmd *cobra.Command, name string, srClient *srsdk.APIClient, ctx context.Context) error {
	configs, _, err := srClient.DefaultApi.GetExporterConfig(ctx, name)
	if err != nil {
		return err
	}

	outputFormat, err := cmd.Flags().GetString("output")
	if err != nil {
		return err
	}

	return output.StructuredOutputForCommand(cmd, outputFormat, configs)
}
