package connect

import (
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/spf13/cobra"
)

func (c *customPluginCommand) newVersionDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe a custom connector plugin version.",
		Args:  cobra.NoArgs,
		RunE:  c.describeVersion,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe custom connector plugin version.",
				Code: "confluent connect custom-plugin version describe --plugin-id ccp-123456 --version-id ver-12345",
			},
		),
	}
	cmd.Flags().String("plugin-id", "", "ID of custom connector plugin.")
	cmd.Flags().String("version-id", "", "ID of custom connector plugin version.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("plugin-id"))
	cobra.CheckErr(cmd.MarkFlagRequired("version-id"))

	return cmd
}

func (c *customPluginCommand) describeVersion(cmd *cobra.Command, args []string) error {
	pluginId, err := cmd.Flags().GetString("plugin-id")
	if err != nil {
		return err
	}
	versionId, err := cmd.Flags().GetString("version-id")
	if err != nil {
		return err
	}

	plugin, err := c.V2Client.DescribeCustomPlugin(pluginId)
	if err != nil {
		return err
	}

	pluginVersion, err := c.V2Client.DescribeCustomPluginVersion(pluginId, versionId)
	if err != nil {
		return err
	}

	return printTableVersion(cmd, plugin, pluginVersion)
}
