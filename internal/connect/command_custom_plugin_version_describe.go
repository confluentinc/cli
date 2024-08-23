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
				Text: `Describe custom connector plugin "ccp-123456" version "ver-12345".`,
				Code: "confluent connect custom-plugin version describe --plugin ccp-123456 --version ver-12345",
			},
		),
	}

	cmd.Flags().String("plugin", "", "ID of custom connector plugin.")
	cmd.Flags().String("version", "", "ID of custom connector plugin version.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("plugin"))
	cobra.CheckErr(cmd.MarkFlagRequired("version"))

	return cmd
}

func (c *customPluginCommand) describeVersion(cmd *cobra.Command, args []string) error {
	plugin, err := cmd.Flags().GetString("plugin")
	if err != nil {
		return err
	}
	version, err := cmd.Flags().GetString("version")
	if err != nil {
		return err
	}

	pluginResp, err := c.V2Client.DescribeCustomPlugin(plugin)
	if err != nil {
		return err
	}

	pluginVersionResp, err := c.V2Client.DescribeCustomPluginVersion(plugin, version)
	if err != nil {
		return err
	}

	return printTableVersion(cmd, pluginResp, pluginVersionResp)
}
