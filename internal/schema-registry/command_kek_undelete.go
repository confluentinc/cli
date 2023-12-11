package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *command) newKekUndeleteCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "undelete <name>",
		Short: "Undelete a KEK.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.kekUndelete,
	}

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

	_, err = client.DescribeKek(args[0], true)
	if err != nil {
		return resource.ResourcesNotFoundError(cmd, resource.Kek, args[0])
	}

	err = client.UndeletKek(args[0])
	if err != nil {
		return err
	}

	output.Printf(c.Config.EnableColor, "Undeleted KEK \"%s\".\n", args[0])
	return nil
}
