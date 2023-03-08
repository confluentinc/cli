package schemaregistry

import (
	"fmt"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/spf13/cobra"
)

func (c *exporterCommand) newDeleteCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "delete <name>",
		Short:       "Delete schema exporter.",
		Args:        cobra.ExactArgs(1),
		RunE:        c.onPremDelete,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
	}

	cmd.Flags().AddFlagSet(pcmd.OnPremSchemaRegistrySet())
	pcmd.AddForceFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *exporterCommand) onPremDelete(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetSrApiClientWithToken(cmd, c.Version, c.AuthToken())
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

	return deleteExporter(args[0], srClient, ctx)
}
