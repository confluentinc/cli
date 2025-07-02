package ccpm

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *pluginCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Custom Connect Plugins.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	cmd.Flags().String("cloud", "", "Filter by cloud provider (AWS, GCP, AZURE).")
	cmd.Flags().String("environment", "", "Environment ID.")
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
