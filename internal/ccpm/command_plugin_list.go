package ccpm

import (
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *pluginCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List custom Connect plugins.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List all custom Connect plugins in an environment.",
				Code: "confluent ccpm plugin list --environment env-12345",
			},
			examples.Example{
				Text: "List custom Connect plugins filtered by cloud provider.",
				Code: "confluent ccpm plugin list --environment env-12345 --cloud AWS",
			},
		),
	}

	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddCloudFlag(cmd)
	pcmd.AddOutputFlag(cmd)
	return cmd
}

func (c *pluginCommand) list(cmd *cobra.Command, args []string) error {
	cloud, err := cmd.Flags().GetString("cloud")
	cloud = strings.ToUpper(cloud)
	if err != nil {
		return err
	}
	environment, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}
	plugins, err := c.V2Client.ListCCPMPlugins(cloud, environment)
	if err != nil {
		return err
	}
	list := output.NewList(cmd)
	for _, plugin := range plugins {
		list.Add(&pluginOut{
			Id:              plugin.GetId(),
			Name:            plugin.Spec.GetDisplayName(),
			Description:     plugin.Spec.GetDescription(),
			Cloud:           plugin.Spec.GetCloud(),
			RuntimeLanguage: plugin.Spec.GetRuntimeLanguage(),
			Environment:     plugin.GetSpec().Environment.GetId(),
		})
	}
	list.Sort(true)
	return list.Print()
}
