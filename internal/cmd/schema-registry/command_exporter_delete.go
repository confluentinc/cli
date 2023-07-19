package schemaregistry

import (
	"context"

	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *command) newExporterDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name-1> <name-2> ... <name-n>",
		Short: "Delete one or more schema exporters.",
		Args:  cobra.MinimumNArgs(1),
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

	if err := c.confirmDeletionExporter(cmd, srClient, ctx, args); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		if _, err := srClient.DefaultApi.DeleteExporter(ctx, id); err != nil {
			return err
		}
		return nil
	}

	deleted, err := resource.Delete(args, deleteFunc, nil)
	resource.PrintDeleteSuccessMsg(deleted, resource.SchemaExporter)

	return err
}

func (c *command) confirmDeletionExporter(cmd *cobra.Command, srClient *srsdk.APIClient, ctx context.Context, args []string) error {
	var name string
	describeFunc := func(id string) error {
		info, _, err := srClient.DefaultApi.GetExporterInfo(ctx, args[0])
		if err == nil && id == args[0] {
			name = info.Name
		}
		return err
	}

	if err := resource.ValidateArgs(pcmd.FullParentName(cmd), args, resource.SchemaExporter, describeFunc); err != nil {
		return err
	}

	if len(args) == 1 {
		if err := form.ConfirmDeletionWithString(cmd, form.DefaultPromptString(resource.SchemaExporter, args[0], name), name); err != nil {
			return err
		}
	} else {
		if ok, err := form.ConfirmDeletionYesNo(cmd, form.DefaultYesNoPromptString(resource.SchemaExporter, args)); err != nil || !ok {
			return err
		}
	}

	return nil
}
