package schemaregistry

import (
	"context"

	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newExporterResumeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resume <name>",
		Short: "Resume schema exporter.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.exporterResume,
	}

	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) exporterResume(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := getApiClient(cmd, c.Config, c.Version)
	if err != nil {
		return err
	}

	return resumeExporter(args[0], srClient, ctx)
}

func resumeExporter(name string, srClient *srsdk.APIClient, ctx context.Context) error {
	if _, _, err := srClient.DefaultApi.ResumeExporter(ctx, name); err != nil {
		return err
	}

	output.Printf(errors.ExporterActionMsg, "Resumed", name)
	return nil
}
