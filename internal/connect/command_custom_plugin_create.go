package connect

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	connectcustompluginv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect-custom-plugin/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type pluginCreateOut struct {
	Id         string `human:"ID" serialized:"id"`
	Name       string `human:"Name" serialized:"name"`
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
				Code: "confluent connect custom-plugin create my-plugin --plugin-file datagen.zip --connector-type source --connector-class io.confluent.kafka.connect.datagen.DatagenConnector",
			},
		),
	}

	cmd.Flags().String("plugin-file", "", "ZIP/JAR custom plugin file.")
	cmd.Flags().String("connector-class", "", "Connector class of custom plugin.")
	cmd.Flags().String("connector-type", "", "Connector type of custom plugin.")
	cmd.Flags().String("description", "", "Description of custom plugin.")
	cmd.Flags().String("documentation-link", "", "Document link of custom plugin.")
	cmd.Flags().StringSlice("sensitive-properties", nil, "A comma-separated list of sensitive property names.")
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

	request := *connectcustompluginv1.NewConnectV1PresignedUrlRequest()
	request.SetContentFormat(extension)

	resp, err := c.V2Client.GetPresignedUrl(request)
	if err != nil {
		return err
	}

	if err := uploadFile(resp.GetUploadUrl(), pluginFileName, resp.GetUploadFormData()); err != nil {
		return err
	}

	createCustomPluginRequest := connectcustompluginv1.ConnectV1CustomConnectorPlugin{
		DisplayName:               connectcustompluginv1.PtrString(displayName),
		Description:               connectcustompluginv1.PtrString(description),
		DocumentationLink:         connectcustompluginv1.PtrString(documentationLink),
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
		Id:   pluginResp.GetId(),
		Name: pluginResp.GetDisplayName(),
	})
	return table.Print()
}
