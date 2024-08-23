package connect

import (
	"fmt"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/deletion"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *customPluginCommand) newVersionDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a custom connector plugin version.",
		Args:  cobra.NoArgs,
		RunE:  c.deleteVersion,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Delete custom connector plugin "ccp-123456" version "ver-123456".`,
				Code: "confluent connect custom-plugin version delete --plugin ccp-123456 --version ver-12345",
			},
		),
	}

	cmd.Flags().String("plugin", "", "ID of custom connector plugin.")
	cmd.Flags().String("version", "", "ID of custom connector plugin version.")
	pcmd.AddForceFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)

	cobra.CheckErr(cmd.MarkFlagRequired("plugin"))
	cobra.CheckErr(cmd.MarkFlagRequired("version"))

	return cmd
}

func (c *customPluginCommand) deleteVersion(cmd *cobra.Command, args []string) error {
	plugin, err := cmd.Flags().GetString("plugin")
	if err != nil {
		return err
	}
	version, err := cmd.Flags().GetString("version")
	if err != nil {
		return err
	}

	_, err = c.V2Client.DescribeCustomPluginVersion(plugin, version)
	if err != nil {
		return err
	}

	existenceFunc := func(id string) bool {
		return id == version || id == plugin
	}

	if err := deletion.ValidateAndConfirmDeletionCustomPluginVersion(cmd, args, existenceFunc, resource.CustomConnectorPlugin, plugin, version); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		return c.V2Client.DeleteCustomPluginVersion(plugin, version)
	}

	_, err = deletion.DeleteWithoutMessage(args, deleteFunc)
	deletedResourceMsg := fmt.Sprintf(`Deleted %s "%s" version "%s".`, resource.CustomConnectorPlugin, plugin, version)
	output.Printf(false, deletedResourceMsg)
	return err
}
