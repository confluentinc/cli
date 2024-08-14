package connect

import (
	"fmt"
	connectcustompluginv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect-custom-plugin/v1"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/utils"
	"github.com/spf13/cobra"
	"path/filepath"
	"strings"
)

type pluginVersionCreateOut struct {
	Version       string `human:"Version" serialized:"version"`
	VersionNumber string `human:"Version Number" serialized:"version_number"`
	IsBeta        string `human:"Is Beta" serialized:"is_beta"`
	ReleaseNotes  string `human:"Release Notes" serialized:"release_notes"`
	ErrorTrace    string `human:"Error Trace,omitempty" serialized:"error_trace,omitempty"`
}

func (c *customPluginCommand) newCreateVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-version <my-plugin-id>",
		Short: "Create a custom connector plugin version.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.createCustomPluginVersion,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create custom connector plugin version for plugin "my-plugin-id".`,
				Code: "confluent connect custom-plugin create-version my-plugin-id --plugin-file datagen.zip --version 0.0.1 --is-beta false release-notes none",
			},
		),
	}

	cmd.Flags().String("plugin-file", "", "ZIP/JAR custom plugin file.")
	cmd.Flags().String("version", "", "Version number for custom plugin version")
	cmd.Flags().String("is-beta", "", "Is Beta flag [true/false] for custom plugin version")
	cmd.Flags().String("release-notes", "", "Release Notes for custom plugin version")
	cmd.Flags().StringSlice("sensitive-properties", nil, "A comma-separated list of sensitive property names.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("plugin-file"))
	cobra.CheckErr(cmd.MarkFlagRequired("version"))
	cobra.CheckErr(cmd.MarkFlagFilename("plugin-file", "zip", "jar"))

	return cmd
}

func (c *customPluginCommand) createCustomPluginVersion(cmd *cobra.Command, args []string) error {
	pluginId := args[0]

	plugin, err := c.V2Client.DescribeCustomPlugin(pluginId)
	if err != nil {
		return err
	}

	pluginVersionFileName, err := cmd.Flags().GetString("plugin-file")
	if err != nil {
		return err
	}

	cloud := plugin.GetCloud()

	version, err := cmd.Flags().GetString("version")
	if err != nil {
		return err
	}
	isBeta, err := cmd.Flags().GetString("is-beta")
	if err != nil {
		return err
	}
	releaseNotes, err := cmd.Flags().GetString("release-notes")
	if err != nil {
		return err
	}
	sensitiveProperties, err := cmd.Flags().GetStringSlice("sensitive-properties")
	if err != nil {
		return err
	}

	extension := strings.ToLower(strings.TrimPrefix(filepath.Ext(pluginVersionFileName), "."))
	if extension != "zip" && extension != "jar" {
		return fmt.Errorf(`only file extensions ".jar" and ".zip" are allowed`)
	}
	cloud = strings.ToUpper(cloud)

	request := connectcustompluginv1.ConnectV1PresignedUrlRequest{
		ContentFormat: connectcustompluginv1.PtrString(extension),
		Cloud:         connectcustompluginv1.PtrString(cloud),
	}

	resp, err := c.V2Client.GetPresignedUrl(request)
	if err != nil {
		return err
	}

	if cloud == "AZURE" {
		if err := utils.UploadFileToAzureBlob(resp.GetUploadUrl(), pluginVersionFileName, strings.ToLower(resp.GetContentFormat())); err != nil {
			return err
		}
	} else {
		if err := utils.UploadFile(resp.GetUploadUrl(), pluginVersionFileName, resp.GetUploadFormData()); err != nil {
			return err
		}
	}

	createCustomPluginVersionRequest := connectcustompluginv1.ConnectV1CustomConnectorPluginVersion{
		Version:                   connectcustompluginv1.PtrString(version),
		IsBeta:                    connectcustompluginv1.PtrString(isBeta),
		ReleaseNotes:              connectcustompluginv1.PtrString(releaseNotes),
		SensitiveConfigProperties: &sensitiveProperties,
		UploadSource: &connectcustompluginv1.ConnectV1CustomConnectorPluginVersionUploadSourceOneOf{
			ConnectV1UploadSourcePresignedUrl: connectcustompluginv1.NewConnectV1UploadSourcePresignedUrl("PRESIGNED_URL_LOCATION", resp.GetUploadId()),
		},
	}

	pluginResp, err := c.V2Client.CreateCustomPluginVersion(createCustomPluginVersionRequest, pluginId)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&pluginVersionCreateOut{
		Version:       pluginResp.GetId(),
		VersionNumber: pluginResp.GetVersion(),
		IsBeta:        pluginResp.GetIsBeta(),
		ReleaseNotes:  pluginResp.GetReleaseNotes(),
	})
	return table.Print()
}
