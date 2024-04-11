package connect

import (
	"fmt"
	"github.com/confluentinc/cli/v3/pkg/ccloudv2"
	"github.com/confluentinc/cli/v3/pkg/utils"
	"github.com/spf13/cobra"
	"strings"

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
				Code: "confluent connect custom-plugin list --cloud AWS",
			},
		),
	}

	c.addListCloudFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *customPluginCommand) list(cmd *cobra.Command, _ []string) error {
	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}

	cloud = strings.ToUpper(cloud)
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

func (c *customPluginCommand) addListCloudFlag(cmd *cobra.Command) {
	cmd.Flags().String("cloud", "", fmt.Sprintf("Specify the cloud provider as %s.", utils.ArrayToCommaDelimitedString(ccloudv2.ByocSupportClouds, "or")))
	pcmd.RegisterFlagCompletionFunc(cmd, "cloud", func(_ *cobra.Command, _ []string) []string { return ccloudv2.ByocSupportClouds })
}
