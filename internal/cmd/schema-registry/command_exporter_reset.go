package schemaregistry

import (
	"context"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *exporterCommand) newResetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset <name>",
		Short: "Reset schema exporter.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.reset,
	}

	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *exporterCommand) reset(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := getApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}

	return resetExporter(args[0], srClient, ctx)
}

func resetExporter(name string, srClient *srsdk.APIClient, ctx context.Context) error {
	if _, _, err := srClient.DefaultApi.ResetExporter(ctx, name); err != nil {
		return err
	}

	utils.Printf(errors.ExporterActionMsg, "Reset", name)
	return nil
}
