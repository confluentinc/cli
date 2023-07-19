package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *command) newExporterDeleteCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "delete <name-1> <name-2> ... <name-n>",
		Short:       "Delete one or more schema exporters.",
		Args:        cobra.MinimumNArgs(1),
		RunE:        c.exporterDeleteOnPrem,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
	}

	cmd.Flags().AddFlagSet(pcmd.OnPremSchemaRegistrySet())
	pcmd.AddForceFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) exporterDeleteOnPrem(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetSrApiClientWithToken(cmd, c.Version, c.AuthToken())
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
