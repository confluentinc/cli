package connect

import (
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/spf13/cobra"
)

type customPluginOutList struct {
	Id   string `human:"ID" serialized:"id"`
	Name string `human:"Name" serialized:"name"`
}

func (c *customPluginCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List custom connector plugins",
		Args:  cobra.NoArgs,
		RunE:  c.list,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List custom connector plugins in the org",
				Code: "confluent connect custom-plugin list",
			},
		),
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *customPluginCommand) list(cmd *cobra.Command, _ []string) error {
	plugins, err := c.V2Client.ListCustomPlugins()
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, plugin := range plugins.GetData() {
		list.Add(&customPluginOutList{
			Name: plugin.GetDisplayName(),
			Id:   plugin.GetId(),
		})
	}
	return list.Print()
}
