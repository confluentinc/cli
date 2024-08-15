package connect

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type customPluginVersionOutList struct {
	Version             string   `human:"Version" serialized:"version"`
	VersionNumber       string   `human:"Version Number" serialized:"version_number"`
	IsBeta              string   `human:"Beta" serialized:"is_beta"`
	ReleaseNotes        string   `human:"Release Notes" serialized:"release_notes"`
	SensitiveProperties []string `human:"Sensitive Properties" serialized:"sensitive_properties"`
}

func (c *customPluginCommand) newVersionListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List custom connector plugin versions for plugin.",
		Args:  cobra.NoArgs,
		RunE:  c.listVersions,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List custom connector plugin versions for plugin",
				Code: "confluent connect custom-plugin version list --plugin-id plugin123",
			},
		),
	}
	cmd.Flags().String("plugin-id", "", "ID of custom connector plugin.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("plugin-id"))

	return cmd
}

func (c *customPluginCommand) listVersions(cmd *cobra.Command, args []string) error {
	pluginId, err := cmd.Flags().GetString("plugin-id")
	if err != nil {
		return err
	}

	versions, err := c.V2Client.ListCustomPluginVersions(pluginId)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, pluginVersion := range versions.Data {
		list.Add(&customPluginVersionOutList{
			Version:             pluginVersion.GetId(),
			VersionNumber:       pluginVersion.GetVersion(),
			IsBeta:              pluginVersion.GetIsBeta(),
			ReleaseNotes:        pluginVersion.GetReleaseNotes(),
			SensitiveProperties: pluginVersion.GetSensitiveConfigProperties(),
		})
	}
	return list.Print()
}
