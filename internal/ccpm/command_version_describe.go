package ccpm

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *customConnectPluginVersionCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <id>",
		Short: "Describe a custom Connect plugin version.",
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
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)
	cobra.CheckErr(cmd.MarkFlagRequired("plugin"))

	return cmd
}

func (c *customConnectPluginVersionCommand) describeVersion(cmd *cobra.Command, args []string) error {
	versionId := args[0]

	pluginId, err := cmd.Flags().GetString("plugin")
	if err != nil {
		return err
	}

	environment, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}
	pluginResp, httpResp, err := c.V2Client.GetCcpmCustomConnectPlugin(pluginId, environment)
	if err != nil {
		return errors.CatchCCloudV2Error(err, httpResp)
	}

	// Use V2Client to call CCPM API
	version, httpResp, err := c.V2Client.GetCcpmCustomConnectPluginVersion(pluginId, versionId, environment)
	if err != nil {
		return errors.CatchCCloudV2Error(err, httpResp)
	}
	return c.printVersionTable(cmd, pluginResp, version)
}
