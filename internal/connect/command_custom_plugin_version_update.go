package connect

import (
	"fmt"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/spf13/cobra"
	"strconv"

	connectcustompluginv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect-custom-plugin/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/resource"
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
				Code: "confluent connect custom-plugin version update --plugin-id plugin123 --version-id ver123 --version 0.0.1 --is-beta false",
			},
		),
	}

	cmd.Flags().String("plugin-id", "", "ID of custom connector plugin.")
	cmd.Flags().String("version-id", "", "ID of custom connector plugin version.")
	cmd.Flags().String("version", "", "Version number for custom plugin version.")
	cmd.Flags().Bool("is-beta", false, "Is Beta flag [true/false] for custom plugin version.")
	cmd.Flags().String("release-notes", "", "Release Notes for custom plugin version.")
	cmd.Flags().StringSlice("sensitive-properties", nil, "A comma-separated list of sensitive property names.")
	pcmd.AddContextFlag(cmd, c.CLICommand)

	cobra.CheckErr(cmd.MarkFlagRequired("plugin-id"))
	cobra.CheckErr(cmd.MarkFlagRequired("version-id"))
	cmd.MarkFlagsOneRequired("version", "is-beta", "release-notes", "sensitive-properties")

	return cmd
}

func (c *customPluginCommand) updateVersion(cmd *cobra.Command, args []string) error {
	pluginId, err := cmd.Flags().GetString("plugin-id")
	if err != nil {
		return err
	}
	versionId, err := cmd.Flags().GetString("version-id")
	if err != nil {
		return err
	}

	updateCustomPluginVersionRequest := connectcustompluginv1.ConnectV1CustomConnectorPluginVersion{}

	if cmd.Flags().Changed("version") {
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

	if _, err := c.V2Client.UpdateCustomPluginVersion(pluginId, versionId, updateCustomPluginVersionRequest); err != nil {
		return err
	}

	outputMessage := fmt.Sprintf("%s\" version \"%s", pluginId, versionId)
	output.Printf(c.Config.EnableColor, errors.UpdatedResourceMsg, resource.CustomConnectorPlugin, outputMessage)
	return nil
}
