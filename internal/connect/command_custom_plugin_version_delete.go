package connect

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/resource"
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

	if _, err := c.V2Client.DescribeCustomPluginVersion(plugin, version); err != nil {
		return err
	}

	existenceFunc := func(id string) bool {
		return id == version || id == plugin
	}

	args = []string{version}
	if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, resource.CustomConnectorPluginVersion); err != nil {
		return err
	}

	if err := c.V2Client.DeleteCustomPluginVersion(plugin, version); err != nil {
		return err
	}

	deletedResourceMsg := "Deleted %s \"%s\" version \"%s\".\n"
	output.Printf(false, deletedResourceMsg, resource.CustomConnectorPlugin, plugin, version)
	return nil
}
