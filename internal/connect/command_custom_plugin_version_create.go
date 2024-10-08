package connect

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	connectcustompluginv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect-custom-plugin/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

type pluginVersionOut struct {
	Version             string   `human:"Version" serialized:"version"`
	VersionNumber       string   `human:"Version Number" serialized:"version_number"`
	IsBeta              string   `human:"Beta" serialized:"is_beta"`
	ReleaseNotes        string   `human:"Release Notes" serialized:"release_notes"`
	SensitiveProperties []string `human:"Sensitive Properties" serialized:"sensitive_properties"`
}

func (c *customPluginCommand) newVersionCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a custom connector plugin version.",
		Args:  cobra.NoArgs,
		RunE:  c.createCustomPluginVersion,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create custom connector plugin version for plugin "ccp-123456".`,
				Code: "confluent connect custom-plugin version create --plugin ccp-123456 --plugin-file datagen.zip --version-number 0.0.1 --beta=true --sensitive-properties passwords,keys,tokens",
			},
		),
	}

	cmd.Flags().String("plugin", "", "ID of custom connector plugin.")
	cmd.Flags().String("plugin-file", "", "Custom plugin ZIP or JAR file.")
	cmd.Flags().String("version-number", "", "Version number of custom plugin version.")
	cmd.Flags().Bool("beta", false, "Mark the custom plugin version as beta.")
	cmd.Flags().String("release-notes", "", "Release notes for custom plugin version.")
	cmd.Flags().StringSlice("sensitive-properties", nil, "A comma-separated list of sensitive property names.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("plugin"))
	cobra.CheckErr(cmd.MarkFlagRequired("plugin-file"))
	cobra.CheckErr(cmd.MarkFlagRequired("version-number"))
	cobra.CheckErr(cmd.MarkFlagFilename("plugin-file", "zip", "jar"))

	return cmd
}

func (c *customPluginCommand) createCustomPluginVersion(cmd *cobra.Command, args []string) error {
	plugin, err := cmd.Flags().GetString("plugin")
	if err != nil {
		return err
	}

	pluginResp, err := c.V2Client.DescribeCustomPlugin(plugin)
	if err != nil {
		return err
	}

	pluginFile, err := cmd.Flags().GetString("plugin-file")
	if err != nil {
		return err
	}

	cloud := pluginResp.GetCloud()

	versionNumber, err := cmd.Flags().GetString("version-number")
	if err != nil {
		return err
	}

	beta, err := cmd.Flags().GetBool("beta")
	if err != nil {
		return err
	}
	isBetaString := strconv.FormatBool(beta)

	releaseNotes, err := cmd.Flags().GetString("release-notes")
	if err != nil {
		return err
	}
	sensitiveProperties, err := cmd.Flags().GetStringSlice("sensitive-properties")
	if err != nil {
		return err
	}

	extension := strings.ToLower(strings.TrimPrefix(filepath.Ext(pluginFile), "."))
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
		if err := utils.UploadFileToAzureBlob(resp.GetUploadUrl(), pluginFile, strings.ToLower(resp.GetContentFormat())); err != nil {
			return err
		}
	} else {
		if err := utils.UploadFile(resp.GetUploadUrl(), pluginFile, resp.GetUploadFormData()); err != nil {
			return err
		}
	}

	createCustomPluginVersionRequest := connectcustompluginv1.ConnectV1CustomConnectorPluginVersion{
		Version:                   connectcustompluginv1.PtrString(versionNumber),
		IsBeta:                    connectcustompluginv1.PtrString(isBetaString),
		ReleaseNotes:              connectcustompluginv1.PtrString(releaseNotes),
		SensitiveConfigProperties: &sensitiveProperties,
		UploadSource: &connectcustompluginv1.ConnectV1CustomConnectorPluginVersionUploadSourceOneOf{
			ConnectV1UploadSourcePresignedUrl: connectcustompluginv1.NewConnectV1UploadSourcePresignedUrl("PRESIGNED_URL_LOCATION", resp.GetUploadId()),
		},
	}

	pluginVersionResp, err := c.V2Client.CreateCustomPluginVersion(createCustomPluginVersionRequest, plugin)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&pluginVersionOut{
		Version:             pluginVersionResp.GetId(),
		VersionNumber:       pluginVersionResp.GetVersion(),
		IsBeta:              pluginVersionResp.GetIsBeta(),
		ReleaseNotes:        pluginVersionResp.GetReleaseNotes(),
		SensitiveProperties: pluginVersionResp.GetSensitiveConfigProperties(),
	})
	return table.Print()
}
