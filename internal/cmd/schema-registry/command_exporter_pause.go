package schemaregistry

import (
	"context"

	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newExporterPauseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pause <name>",
		Short: "Pause schema exporter.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.exporterPause,
	}

	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) exporterPause(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := getApiClient(cmd, c.Config, c.Version)
	if err != nil {
		return err
	}

	return pauseExporter(args[0], srClient, ctx)
}

func pauseExporter(name string, srClient *srsdk.APIClient, ctx context.Context) error {
	if _, _, err := srClient.DefaultApi.PauseExporter(ctx, name); err != nil {
		return err
	}

	output.Printf(errors.ExporterActionMsg, "Paused", name)
	return nil
}
