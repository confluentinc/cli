package connect

import (
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/spf13/cobra"
)

func (c *customPluginCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <id>",
		Short: "Describe a custom connector plugin.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.describe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe custom connector plugin",
				Code: "confluent connect custom-plugin describe ccp-123456",
			},
		),
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *customPluginCommand) describe(cmd *cobra.Command, args []string) error {
	plugin, err := c.V2Client.DescribeCustomPlugin(args[0])
	if err != nil {
		return err
	}
	out := &customPluginOut{
		Id:                        plugin.GetId(),
		Name:                      plugin.GetDisplayName(),
		Description:               plugin.GetDescription(),
		ConnectorClass:            plugin.GetConnectorClass(),
		ConnectorType:             plugin.GetConnectorType(),
		SensitiveConfigProperties: plugin.GetSensitiveConfigProperties(),
	}
	if output.GetFormat(cmd) == output.Human {
		output.Println("Plugin Details")
		table := output.NewTable(cmd)
		table.Add(out)
		return table.Print()
	}
	return output.SerializedOutput(cmd, out)
}
