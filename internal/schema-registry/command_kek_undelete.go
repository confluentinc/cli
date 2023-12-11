package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/deletion"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *command) newKekUndeleteCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "undelete <name-1> [name-2] ... [name-n]",
		Short: "Undelete one or more KEKs.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.kekUndelete,
	}

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

func (c *command) kekUndelete(cmd *cobra.Command, args []string) error {
	client, err := c.GetSchemaRegistryClient(cmd)
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

	if err := deletion.ValidateAndConfirmUndeletion(cmd, args, existenceFunc, resource.Kek, kek.GetName()); err != nil {
		return err
	}

	undeleteFunc := func(name string) error {
		return client.UndeleteKek(name)
	}

	_, err = deletion.Undelete(args, undeleteFunc, resource.Kek)
	return err
}
