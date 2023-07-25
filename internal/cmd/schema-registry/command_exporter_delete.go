package schemaregistry

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *command) newExporterDeleteCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a schema exporter.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.exporterDelete,
	}

	pcmd.AddForceFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		cmd.Flags().AddFlagSet(pcmd.OnPremSchemaRegistrySet())
	}
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) exporterDelete(cmd *cobra.Command, args []string) error {
	client, err := c.GetSchemaRegistryClient()
	if err != nil {
		return err
	}

	info, err := client.GetExporterInfo(args[0])
	if err != nil {
		return err
	}

	promptMsg := fmt.Sprintf(errors.DeleteResourceConfirmMsg, resource.SchemaExporter, info.Name, info.Name)
	if _, err := form.ConfirmDeletion(cmd, promptMsg, info.Name); err != nil {
		return err
	}

	if err := client.DeleteExporter(args[0]); err != nil {
		return err
	}

	output.Printf(errors.DeletedResourceMsg, resource.SchemaExporter, args[0])
	return nil
}
