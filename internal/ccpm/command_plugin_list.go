package ccpm

import (
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

	cmd.Flags().String("environment", "", "Environment ID.")
	cmd.Flags().String("cloud", "", "Filter by cloud provider (AWS, GCP, AZURE).")
	cobra.CheckErr(cmd.MarkFlagRequired("environment"))
	pcmd.AddOutputFlag(cmd)
	return cmd
}

func (c *pluginCommand) list(cmd *cobra.Command, args []string) error {
	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}
	environment, err := cmd.Flags().GetString("environment")
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
			Id:              *plugin.Id,
			Name:            *plugin.Spec.DisplayName,
			Description:     *plugin.Spec.Description,
			Cloud:           *plugin.Spec.Cloud,
			RuntimeLanguage: *plugin.Spec.RuntimeLanguage,
			Environment:     plugin.Spec.Environment.Id,
		})
	}
	list.Sort(true)
	return list.Print()
}
