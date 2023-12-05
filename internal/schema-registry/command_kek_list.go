package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type kekListOut struct {
	Name string `human:"Name" serialized:"name"`
}

func (c *command) newKekListCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Kek.",
		Args:  cobra.NoArgs,
		RunE:  c.kekList,
	}

	cmd.Flags().Bool("deleted", false, "Include deleted Kek.")

	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		addCaLocationFlag(cmd)
		addSchemaRegistryEndpointFlag(cmd) // guess it's needed?
	}
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) kekList(cmd *cobra.Command, _ []string) error {
	client, err := c.GetSchemaRegistryClient(cmd)
	if err != nil {
		return err
	}

	deleted, err := cmd.Flags().GetBool("deleted")
	if err != nil {
		return err
	}

	keks, err := client.ListKeks(deleted)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, kek := range keks {
		list.Add(&kekListOut{Name: kek})
	}
	return list.Print()
}
