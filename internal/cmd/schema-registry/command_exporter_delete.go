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

	if confirm, err := c.confirmDeletionExporter(cmd, srClient, ctx, args); err != nil {
		return err
	} else if !confirm {
		return nil
	}

	deleteFunc := func(id string) error {
		if _, err := srClient.DefaultApi.DeleteExporter(ctx, id); err != nil {
			return err
		}
		return nil
	}

	deleted, err := resource.Delete(args, deleteFunc)
	resource.PrintDeleteSuccessMsg(deleted, resource.SchemaExporter)

	return err
}

func (c *command) confirmDeletionExporter(cmd *cobra.Command, srClient *srsdk.APIClient, ctx context.Context, args []string) (bool, error) {
	var name string
	describeFunc := func(id string) error {
		info, _, err := srClient.DefaultApi.GetExporterInfo(ctx, args[0])
		if err != nil {
			return err
		}
		if id == args[0] {
			name = info.Name
		}

		return nil
	}

	if err := resource.ValidateArgs(pcmd.FullParentName(cmd), args, resource.SchemaExporter, describeFunc); err != nil {
		return false, err
	}

	if len(args) > 1 {
		return form.ConfirmDeletionYesNo(cmd, form.DefaultYesNoPromptString(resource.SchemaExporter, args))
	}

	if err := form.ConfirmDeletionWithString(cmd, form.DefaultPromptString(resource.SchemaExporter, args[0], name), name); err != nil {
		return false, err
	}

	return true, nil
}
