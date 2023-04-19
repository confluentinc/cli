package schemaregistry

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *command) newExporterDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete schema exporter.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.exporterDelete,
	}

	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)
	pcmd.AddForceFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) exporterDelete(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := getApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}

	info, _, err := srClient.DefaultApi.GetExporterInfo(ctx, args[0])
	if err != nil {
		return err
	}

	promptMsg := fmt.Sprintf(errors.DeleteResourceConfirmMsg, resource.SchemaExporter, info.Name, info.Name)
	if err := form.ConfirmDeletionTypeCustomPrompt(cmd, promptMsg, info.Name); err != nil {
		return err
	}

	return deleteExporter(args[0], srClient, ctx)
}

func deleteExporter(name string, srClient *srsdk.APIClient, ctx context.Context) error {
	if _, err := srClient.DefaultApi.DeleteExporter(ctx, name); err != nil {
		return err
	}

	output.Printf(errors.DeletedResourceMsg, resource.SchemaExporter, name)
	return nil
}
