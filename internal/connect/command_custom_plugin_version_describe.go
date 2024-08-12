package connect

import (
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/spf13/cobra"
)

func (c *customPluginCommand) newDescribeVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe-version <plugin-id> <version-id",
		Short: "Describe a custom connector plugin version.",
		Args:  cobra.ExactArgs(2),
		RunE:  c.describeVersion,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe custom connector plugin version",
				Code: "confluent connect custom-plugin describe-version ccp-123456 ver-12345",
			},
		),
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)
	return cmd
}

func (c *customPluginCommand) describeVersion(cmd *cobra.Command, args []string) error {
	plugin, err := c.V2Client.DescribeCustomPlugin(args[0])
	if err != nil {
		return err
	}

	pluginVersion, err := c.V2Client.DescribeCustomPluginVersion(args[0], args[1])
	if err != nil {
		return err
	}

	return printTableVersion(cmd, plugin, pluginVersion)
}
