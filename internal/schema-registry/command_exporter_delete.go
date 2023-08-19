package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/deletion"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *command) newExporterDeleteCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name-1> <name-2> ... <name-n>",
		Short: "Delete one or more schema exporters.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.exporterDelete,
	}

	pcmd.AddForceFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		addCaLocationFlag(cmd)
		addSchemaRegistryEndpointFlag(cmd)
	}
	pcmd.AddOutputFlag(cmd)

	if cfg.IsCloudLogin() {
		// Deprecated
		pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
		cobra.CheckErr(cmd.Flags().MarkHidden("api-key"))

		// Deprecated
		pcmd.AddApiSecretFlag(cmd)
		cobra.CheckErr(cmd.Flags().MarkHidden("api-secret"))
	}

	return cmd
}

func (c *command) exporterDelete(cmd *cobra.Command, args []string) error {
	client, err := c.GetSchemaRegistryClient(cmd)
	if err != nil {
		return err
	}

	info, err := client.GetExporterInfo(args[0])
	if err != nil {
		return resource.ResourcesNotFoundError(cmd, resource.SchemaExporter, args[0])
	}

	existenceFunc := func(id string) bool {
		_, err := client.GetExporterInfo(id)
		return err == nil
	}

	if confirm, err := deletion.ValidateAndConfirmDeletion(cmd, args, existenceFunc, resource.SchemaExporter, info.Name); err != nil {
		return err
	} else if !confirm {
		return nil
	}

	deleteFunc := func(id string) error {
		return client.DeleteExporter(id)
	}

	_, err = deletion.Delete(args, deleteFunc, resource.SchemaExporter)
	return err
}
