package ccpm

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *pluginCommand) newDescribeVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <version-id>",
		Short: "Describe a Custom Connect Plugin Version.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.describeVersion,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe a specific version of a custom connect plugin.",
				Code: "confluent ccpm plugin version describe version-789012 --plugin plugin-123456 --environment env-abcdef",
			},
			examples.Example{
				Text: "Get detailed information about version 1.0.0 of a plugin.",
				Code: "confluent ccpm plugin version describe version-1.0.0 --plugin plugin-123456 --environment env-abcdef",
			},
		),
	}

	cmd.Flags().String("plugin", "", "Plugin ID.")
	cmd.Flags().String("environment", "", "Environment ID.")
	cobra.CheckErr(cmd.MarkFlagRequired("plugin"))
	cobra.CheckErr(cmd.MarkFlagRequired("environment"))

	return cmd
}

func (c *pluginCommand) describeVersion(cmd *cobra.Command, args []string) error {
	versionId := args[0]

	pluginId, err := cmd.Flags().GetString("plugin")
	if err != nil {
		return err
	}

	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}
	pluginResp, err := c.V2Client.DescribeCCPMPlugin(pluginId, environment)
	if err != nil {
		return err
	}

	// Use V2Client to call CCPM API
	version, err := c.V2Client.DescribeCCPMPluginVersion(pluginId, versionId, environment)
	if err != nil {
		return err
	}
	return c.printVersionTable(cmd, pluginResp, version)
}
