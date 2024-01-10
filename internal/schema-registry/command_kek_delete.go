package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/deletion"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *command) newKekDeleteCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name-1> [name-2] ... [name-n]",
		Short: "Delete one or more Key Encryption Keys (KEKs).",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.kekDelete,
	}

	cmd.Flags().Bool("permanent", false, "Delete the Key Encryption Key (KEK) permanently.")
	pcmd.AddForceFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		addCaLocationFlag(cmd)
		addSchemaRegistryEndpointFlag(cmd)
	}

	return cmd
}

func (c *command) kekDelete(cmd *cobra.Command, args []string) error {
	client, err := c.GetSchemaRegistryClient(cmd)
	if err != nil {
		return err
	}

	permanent, err := cmd.Flags().GetBool("permanent")
	if err != nil {
		return err
	}

	kek, err := client.DescribeKek(args[0], true)
	if err != nil {
		return resource.ResourcesNotFoundError(cmd, resource.Kek, args[0])
	}

	existenceFunc := func(name string) bool {
		_, err := client.DescribeKek(name, true)
		return err == nil
	}

	if err := deletion.ValidateAndConfirmDeletion(cmd, args, existenceFunc, resource.Kek, kek.GetName()); err != nil {
		return err
	}

	deleteFunc := func(name string) error {
		return client.DeleteKek(name, permanent)
	}

	_, err = deletion.Delete(args, deleteFunc, resource.Kek)
	return err
}
