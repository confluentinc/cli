package schemaregistry

import (
	"context"
	"fmt"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *exporterCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete schema exporter.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.delete,
	}

	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)
	pcmd.AddForceFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *exporterCommand) delete(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := getApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}

	info, _, err := srClient.DefaultApi.GetExporterInfo(ctx, args[0])
	if err != nil {
		return err
	}

	promptMsg := fmt.Sprintf(errors.DeleteResourceConfirmMsg, resource.SchemaExporter, info.Name, info.Name)
	if _, err := form.ConfirmDeletion(cmd, promptMsg, info.Name); err != nil {
		return err
	}

	return deleteExporter(cmd, args[0], srClient, ctx)
}

func deleteExporter(cmd *cobra.Command, name string, srClient *srsdk.APIClient, ctx context.Context) error {
	if _, err := srClient.DefaultApi.DeleteExporter(ctx, name); err != nil {
		return err
	}

	output.Printf(errors.DeletedResourceMsg, resource.SchemaExporter, name)
	return nil
}
