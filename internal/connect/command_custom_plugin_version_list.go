package connect

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *customPluginCommand) newVersionListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List custom connector plugin versions for plugin.",
		Args:  cobra.NoArgs,
		RunE:  c.listVersions,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List custom connector plugin versions for plugin",
				Code: "confluent connect custom-plugin version list --plugin ccp-123456",
			},
		),
	}
	cmd.Flags().String("plugin", "", "ID of custom connector plugin.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("plugin"))

	return cmd
}

func (c *customPluginCommand) listVersions(cmd *cobra.Command, args []string) error {
	plugin, err := cmd.Flags().GetString("plugin")
	if err != nil {
		return err
	}

	versionsResp, err := c.V2Client.ListCustomPluginVersions(plugin)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, pluginVersion := range versionsResp.Data {
		list.Add(&pluginVersionOut{
			Version:             pluginVersion.GetId(),
			VersionNumber:       pluginVersion.GetVersion(),
			IsBeta:              pluginVersion.GetIsBeta(),
			ReleaseNotes:        pluginVersion.GetReleaseNotes(),
			SensitiveProperties: pluginVersion.GetSensitiveConfigProperties(),
		})
	}
	return list.Print()
}
