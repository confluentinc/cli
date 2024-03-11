package connect

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type customPluginOutList struct {
	Id    string `human:"ID" serialized:"id"`
	Name  string `human:"Name" serialized:"name"`
	Cloud string `human:"Cloud" serialized:"cloud"`
}

func (c *customPluginCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List custom connector plugins.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List custom connector plugins in the org",
				Code: "confluent connect custom-plugin list --cloud aws",
			},
		),
	}

	cmd.Flags().String("cloud", "", "Filter plugins by cloud provider.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *customPluginCommand) list(cmd *cobra.Command, _ []string) error {
	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}

	plugins, err := c.V2Client.ListCustomPlugins(cloud)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, plugin := range plugins {
		list.Add(&customPluginOutList{
			Name:  plugin.GetDisplayName(),
			Id:    plugin.GetId(),
			Cloud: plugin.GetCloud(),
		})
	}
	return list.Print()
}
