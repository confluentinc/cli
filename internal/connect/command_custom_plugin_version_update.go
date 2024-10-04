package connect

import (
	"strconv"

	"github.com/spf13/cobra"

	connectcustompluginv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect-custom-plugin/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *customPluginCommand) newVersionUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a custom connector plugin version metadata.",
		Args:  cobra.NoArgs,
		RunE:  c.updateVersion,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update custom connector plugin version number, beta and sensitive properties for plugin "ccp-123456" version "ver-12345."`,
				Code: "confluent connect custom-plugin version update --plugin ccp-123456 --version ver-12345 --version-number 0.0.1 --beta=true --sensitive-properties passwords,keys,tokens",
			},
			examples.Example{
				Text: `Update release notes for custom connector plugin "ccp-123456" version "ver-12345."`,
				Code: `confluent connect custom-plugin version update --plugin ccp-123456 --version ver-12345 --release-notes "New release."`,
			},
		),
	}

	cmd.Flags().String("plugin", "", "ID of custom connector plugin.")
	cmd.Flags().String("version", "", "ID of custom connector plugin version.")
	cmd.Flags().String("version-number", "", "Version number of custom plugin version.")
	cmd.Flags().Bool("beta", false, "Mark the custom plugin version as beta.")
	cmd.Flags().String("release-notes", "", "Release notes for custom plugin version.")
	cmd.Flags().StringSlice("sensitive-properties", nil, "A comma-separated list of sensitive property names.")
	pcmd.AddContextFlag(cmd, c.CLICommand)

	cobra.CheckErr(cmd.MarkFlagRequired("plugin"))
	cobra.CheckErr(cmd.MarkFlagRequired("version"))
	cmd.MarkFlagsOneRequired("version-number", "beta", "release-notes", "sensitive-properties")

	return cmd
}

func (c *customPluginCommand) updateVersion(cmd *cobra.Command, args []string) error {
	plugin, err := cmd.Flags().GetString("plugin")
	if err != nil {
		return err
	}
	version, err := cmd.Flags().GetString("version")
	if err != nil {
		return err
	}

	updateCustomPluginVersionRequest := connectcustompluginv1.ConnectV1CustomConnectorPluginVersion{}

	if cmd.Flags().Changed("version-number") {
		if versionNumber, err := cmd.Flags().GetString("version-number"); err != nil {
			return err
		} else {
			updateCustomPluginVersionRequest.SetVersion(versionNumber)
		}
	}
	if cmd.Flags().Changed("beta") {
		beta, err := cmd.Flags().GetBool("beta")
		if err != nil {
			return err
		}
		isBetaString := strconv.FormatBool(beta)
		updateCustomPluginVersionRequest.SetIsBeta(isBetaString)
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

	if pluginResp, err := c.V2Client.UpdateCustomPluginVersion(plugin, version, updateCustomPluginVersionRequest); err != nil {
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
