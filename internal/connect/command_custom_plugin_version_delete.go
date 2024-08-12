package connect

import (
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/deletion"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *customPluginCommand) newDeleteVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-version <plugin-id> <version-id",
		Short: "Delete a custom connector plugin version.",
		Args:  cobra.ExactArgs(2),
		RunE:  c.deleteVersion,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "delete-version custom connector plugin version",
				Code: "confluent connect custom-plugin delete ccp-123456 ver-12345",
			},
		),
	}

	pcmd.AddForceFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)

	return cmd
}

func (c *customPluginCommand) deleteVersion(cmd *cobra.Command, args []string) error {
	version, err := c.V2Client.DescribeCustomPluginVersion(args[0], args[1])
	if err != nil {
		return err
	}

	existenceFunc := func(id string) bool {
		return version.GetId() == args[1]
	}

	if err := deletion.ValidateAndConfirmDeletion(cmd, args, existenceFunc, resource.CustomConnectorPlugin, args[1]); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		return c.V2Client.DeleteCustomPluginVersion(args[0], args[1])
	}

	_, err = deletion.Delete(args, deleteFunc, resource.CustomConnectorPlugin)
	return err
}
