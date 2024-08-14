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
	IsBeta              string   `human:"Is Beta" serialized:"is_beta"`
	ReleaseNotes        string   `human:"Release Notes" serialized:"release_notes"`
	SensitiveProperties []string `human:"Sensitive Properties" serialized:"sensitive_properties"`
}

func (c *customPluginCommand) newListVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-versions <plugin-id>",
		Short: "List custom connector plugin versions for plugin.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.listVersions,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List custom connector plugin versions for plugin",
				Code: "confluent connect custom-plugin list-versions <plugin-id>",
			},
		),
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *customPluginCommand) listVersions(cmd *cobra.Command, args []string) error {
	versions, err := c.V2Client.ListCustomPluginVersions(args[0])
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
