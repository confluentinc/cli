package connect

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	pluginFields          = []string{"pluginName", "pluginType"}
	pluginHumanFields     = []string{"Plugin Name", "Type"}
	pluginStructureLabels = []string{"plugin_name", "type"}
)

func (c *pluginCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List connector plugin types.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List connectors in the current or specified Kafka cluster context.",
				Code: "confluent connect plugin list",
			},
			examples.Example{
				Code: "confluent connect plugin list --cluster lkc-123456",
			},
		),
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *pluginCommand) list(cmd *cobra.Command, _ []string) error {
	plugins, _, err := c.getPlugins()
	if err != nil {
		return err
	}

	outputWriter, err := output.NewListOutputWriter(cmd, pluginFields, pluginHumanFields, pluginStructureLabels)
	if err != nil {
		return err
	}

	for _, plugin := range plugins {
		pluginListDisplay := &pluginListDisplay{
			pluginName: plugin.Class,
			pluginType: plugin.Type,
		}
		outputWriter.AddElement(pluginListDisplay)
	}
	outputWriter.StableSort()

	return outputWriter.Out()
}
