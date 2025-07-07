package connect

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	connectcustompluginv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect-custom-plugin/v1"

	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

type pluginCreateOut struct {
	Id         string `human:"ID" serialized:"id"`
	Name       string `human:"Name" serialized:"name"`
	Cloud      string `human:"Cloud" serialized:"cloud"`
	ErrorTrace string `human:"Error Trace,omitempty" serialized:"error_trace,omitempty"`
}

func (c *customPluginCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a custom connector plugin.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.createCustomPlugin,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create custom connector plugin "my-plugin".`,
				Code: "confluent connect custom-plugin create my-plugin --plugin-file datagen.zip --connector-type source --connector-class io.confluent.kafka.connect.datagen.DatagenConnector --cloud aws",
			},
		),
	}

	cmd.Flags().String("plugin-file", "", "ZIP/JAR custom plugin file.")
	cmd.Flags().String("connector-class", "", "Connector class of custom plugin.")
	cmd.Flags().String("connector-type", "", "Connector type of custom plugin.")
	cmd.Flags().String("description", "", "Description of custom plugin.")
	cmd.Flags().String("documentation-link", "", "Document link of custom plugin.")
	cmd.Flags().StringSlice("sensitive-properties", nil, "A comma-separated list of sensitive property names.")
	c.addCloudFlag(cmd, "aws")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("plugin-file"))
	cobra.CheckErr(cmd.MarkFlagRequired("connector-class"))
	cobra.CheckErr(cmd.MarkFlagRequired("connector-type"))
	cobra.CheckErr(cmd.MarkFlagFilename("plugin-file", "zip", "jar"))

	return cmd
}

func (c *customPluginCommand) createCustomPlugin(cmd *cobra.Command, args []string) error {
	displayName := args[0]
	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}
	documentationLink, err := cmd.Flags().GetString("documentation-link")
	if err != nil {
		return err
	}
	pluginFileName, err := cmd.Flags().GetString("plugin-file")
	if err != nil {
		return err
	}

	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}
	connectorClass, err := cmd.Flags().GetString("connector-class")
	if err != nil {
		return err
	}
	connectorType, err := cmd.Flags().GetString("connector-type")
	if err != nil {
		return err
	}
	sensitiveProperties, err := cmd.Flags().GetStringSlice("sensitive-properties")
	if err != nil {
		return err
	}

	extension := strings.ToLower(strings.TrimPrefix(filepath.Ext(pluginFileName), "."))
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
		if err := utils.UploadFileToAzureBlob(resp.GetUploadUrl(), pluginFileName, strings.ToLower(resp.GetContentFormat())); err != nil {
			return err
		}
	} else if cloud == "GCP" {
		if err := utils.UploadFileToGoogleCloudStorage(resp.GetUploadUrl(), pluginFileName, strings.ToLower(resp.GetContentFormat())); err != nil {
			return err
		}
	} else {
		if err := utils.UploadFile(resp.GetUploadUrl(), pluginFileName, resp.GetUploadFormData()); err != nil {
			return err
		}
	}

	createCustomPluginRequest := connectcustompluginv1.ConnectV1CustomConnectorPlugin{
		DisplayName:               connectcustompluginv1.PtrString(displayName),
		Description:               connectcustompluginv1.PtrString(description),
		DocumentationLink:         connectcustompluginv1.PtrString(documentationLink),
		Cloud:                     connectcustompluginv1.PtrString(cloud),
		ConnectorClass:            connectcustompluginv1.PtrString(connectorClass),
		ConnectorType:             connectcustompluginv1.PtrString(connectorType),
		SensitiveConfigProperties: &sensitiveProperties,
		UploadSource: &connectcustompluginv1.ConnectV1CustomConnectorPluginUploadSourceOneOf{
			ConnectV1UploadSourcePresignedUrl: connectcustompluginv1.NewConnectV1UploadSourcePresignedUrl("PRESIGNED_URL_LOCATION", resp.GetUploadId()),
		},
	}

	pluginResp, err := c.V2Client.CreateCustomPlugin(createCustomPluginRequest)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&pluginCreateOut{
		Id:    pluginResp.GetId(),
		Name:  pluginResp.GetDisplayName(),
		Cloud: pluginResp.GetCloud(),
	})
	return table.Print()
}

func (c *customPluginCommand) addCloudFlag(cmd *cobra.Command, value string) {
	cmd.Flags().String("cloud", value, fmt.Sprintf("Specify the cloud provider as %s.", utils.ArrayToCommaDelimitedString(ccloudv2.ByocSupportClouds, "or")))
	pcmd.RegisterFlagCompletionFunc(cmd, "cloud", func(_ *cobra.Command, _ []string) []string { return ccloudv2.ByocSupportClouds })
}
