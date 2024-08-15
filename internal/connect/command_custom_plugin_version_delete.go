package connect

import (
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/deletion"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *customPluginCommand) newDeleteVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <plugin-id> <version-id>",
		Short: "Delete a custom connector plugin version.",
		Args:  cobra.ExactArgs(2),
		RunE:  c.deleteVersion,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Delete custom connector plugin version for plugin "my-plugin-id" version "my-plugin-version".`,
				Code: "confluent connect custom-plugin delete-version ccp-123456 ver-12345",
			},
		),
	}

	pcmd.AddForceFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)

	return cmd
}

func (c *customPluginCommand) deleteVersion(cmd *cobra.Command, args []string) error {
	_, err := c.V2Client.DescribeCustomPluginVersion(args[0], args[1])
	if err != nil {
		return err
	}

	existenceFunc := func(id string) bool {
		return id == args[1] || id == args[0]
	}

	if err := deletion.ValidateAndConfirmDeletionCustomPluginVersion(cmd, args, existenceFunc, resource.CustomConnectorPlugin, args[1]); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		return c.V2Client.DeleteCustomPluginVersion(args[0], args[1])
	}

	_, err = deletion.DeleteWithoutMessage(args, deleteFunc)
	deletedResourceMsg := `Deleted %s version for plugin "` + args[0] + `" version "` + args[1] + `"`
	output.Printf(false, deletedResourceMsg, resource.CustomConnectorPlugin)
	return err
}
