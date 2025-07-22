package ccpm

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *pluginCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <id>",
		Short: "Describe a custom Connect plugin.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.describe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe a custom Connect plugin by ID.",
				Code: "confluent ccpm plugin describe plugin-123456 --environment env-12345",
			},
		),
	}

	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *pluginCommand) describe(cmd *cobra.Command, args []string) error {
	pluginId := args[0]

	environment, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	// Use V2Client to call CCPM API
	plugin, err := c.V2Client.DescribeCCPMPlugin(pluginId, environment)
	if err != nil {
		return err
	}
	return printCustomConnectPluginTable(cmd, plugin)
}
