package connect

import (
	"fmt"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/spf13/cobra"

	connectcustompluginv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect-custom-plugin/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *customPluginCommand) newUpdateVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-version <plugin-id> <version-id>",
		Short: "Update a custom connector plugin version metadata.",
		Args:  cobra.ExactArgs(2),
		RunE:  c.updateVersion,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update custom connector plugin version for plugin "my-plugin-id" version "my-plugin-version."`,
				Code: "confluent connect custom-plugin update-version my-plugin-id my-version-id --version 0.0.1 --is-beta false release-notes none",
			},
		),
	}

	cmd.Flags().String("version", "", "Version number for custom plugin version")
	cmd.Flags().String("is-beta", "", "Is Beta flag [true/false] for custom plugin version")
	cmd.Flags().String("release-notes", "", "Release Notes for custom plugin version")
	cmd.Flags().StringSlice("sensitive-properties", nil, "A comma-separated list of sensitive property names.")
	pcmd.AddContextFlag(cmd, c.CLICommand)

	cmd.MarkFlagsOneRequired("version", "is-beta", "release-notes", "sensitive-properties")

	return cmd
}

func (c *customPluginCommand) updateVersion(cmd *cobra.Command, args []string) error {
	pluginId := args[0]
	versionId := args[1]
	updateCustomPluginVersionRequest := connectcustompluginv1.ConnectV1CustomConnectorPluginVersion{}

	if cmd.Flags().Changed("version") {
		if version, err := cmd.Flags().GetString("version"); err != nil {
			return err
		} else {
			updateCustomPluginVersionRequest.SetVersion(version)
		}
	}
	if cmd.Flags().Changed("is-beta") {
		if isBeta, err := cmd.Flags().GetString("is-beta"); err != nil {
			return err
		} else {
			updateCustomPluginVersionRequest.SetIsBeta(isBeta)
		}
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

	outputMessage := fmt.Sprintf("%s %s", args[0], args[1])
	output.Printf(c.Config.EnableColor, errors.UpdatedResourceMsg, resource.CustomConnectorPlugin, outputMessage)
	return nil
}
