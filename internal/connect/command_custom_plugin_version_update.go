package connect

import (
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/spf13/cobra"
	"strconv"

	connectcustompluginv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect-custom-plugin/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *customPluginCommand) newVersionUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a custom connector plugin version metadata.",
		Args:  cobra.NoArgs,
		RunE:  c.updateVersion,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update custom connector plugin version for plugin "plugin123" version "ver123."`,
				Code: "confluent connect custom-plugin version update --plugin plugin123 --version ver123 --version-number 0.0.1",
			},
		),
	}

	cmd.Flags().String("plugin", "", "ID of custom connector plugin.")
	cmd.Flags().String("version", "", "ID of custom connector plugin version.")
	cmd.Flags().String("version-number", "", "Version number for custom plugin version.")
	cmd.Flags().Bool("is-beta", false, "Is Beta flag for custom plugin version.")
	cmd.Flags().String("release-notes", "", "Release Notes for custom plugin version.")
	cmd.Flags().StringSlice("sensitive-properties", nil, "A comma-separated list of sensitive property names.")
	pcmd.AddContextFlag(cmd, c.CLICommand)

	cobra.CheckErr(cmd.MarkFlagRequired("plugin"))
	cobra.CheckErr(cmd.MarkFlagRequired("version"))
	cmd.MarkFlagsOneRequired("version-number", "is-beta", "release-notes", "sensitive-properties")

	return cmd
}

func (c *customPluginCommand) updateVersion(cmd *cobra.Command, args []string) error {
	pluginId, err := cmd.Flags().GetString("plugin")
	if err != nil {
		return err
	}
	versionId, err := cmd.Flags().GetString("version")
	if err != nil {
		return err
	}

	updateCustomPluginVersionRequest := connectcustompluginv1.ConnectV1CustomConnectorPluginVersion{}

	if cmd.Flags().Changed("version-number") {
		if version, err := cmd.Flags().GetString("version"); err != nil {
			return err
		} else {
			updateCustomPluginVersionRequest.SetVersion(version)
		}
	}
	if cmd.Flags().Changed("is-beta") {
		isBetaBool, err := cmd.Flags().GetBool("is-beta")
		if err != nil {
			return err
		}
		isBeta := strconv.FormatBool(isBetaBool)
		updateCustomPluginVersionRequest.SetIsBeta(isBeta)
	}
	if cmd.Flags().Changed("release-notes") {
		if releaseNotes, err := cmd.Flags().GetString("release-notes"); err != nil {
			return err
		} else {
			updateCustomPluginVersionRequest.SetReleaseNotes(releaseNotes)
		}
	}
	if cmd.Flags().Changed("sensitive-properties") {
		if sensitiveProperties, err := cmd.Flags().GetStringSlice("sensitive-properties"); err != nil {
			return err
		} else {
			updateCustomPluginVersionRequest.SetSensitiveConfigProperties(sensitiveProperties)
		}
	}

	if pluginResp, err := c.V2Client.UpdateCustomPluginVersion(pluginId, versionId, updateCustomPluginVersionRequest); err != nil {
		return err
	} else {
		table := output.NewTable(cmd)
		table.Add(&pluginVersionOut{
			Version:             pluginResp.GetId(),
			VersionNumber:       pluginResp.GetVersion(),
			IsBeta:              pluginResp.GetIsBeta(),
			ReleaseNotes:        pluginResp.GetReleaseNotes(),
			SensitiveProperties: pluginResp.GetSensitiveConfigProperties(),
		})
		return table.Print()
	}
}
