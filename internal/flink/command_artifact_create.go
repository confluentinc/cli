package flink

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
	Name           string `human:"Name" serialized:"name"`
	Id             string `human:"Plugin ID" serialized:"id"`
	ConnectorClass string `human:"Version ID" serialized:"connector_class"`
	ContentFormat  string `human:"Content Format" serialized:"content_format"`
	Scope          string `human:"Scope"`
	ErrorTrace     string `human:"Error Trace,omitempty" serialized:"error_trace,omitempty"`
}

func (c *command) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a flink udf artifact.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.createArtifact,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create flink udf artifact "my-flink-artifact".`,
				Code: "confluent flink artifact create my-flink-artifact --artifact-file /Users/xyz/Documents/config/plugin.jar",
			},
		),
	}

	cmd.Flags().String("artifact-file", "", "ZIP/JAR flink udf artifact file.")
	cmd.Flags().String("description", "", "Description of flink udf artifact.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("artifact-file"))
	cobra.CheckErr(cmd.MarkFlagFilename("artifact-file", "jar"))

	return cmd
}

func (c *command) createArtifact(cmd *cobra.Command, args []string) error {
	displayName := args[0]
	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}
	pluginFileName, err := cmd.Flags().GetString("artifact-file")
	if err != nil {
		return err
	}

	extension := strings.ToLower(strings.TrimPrefix(filepath.Ext(pluginFileName), "."))
	if extension != "jar" {
		return fmt.Errorf(`only file extensions ".jar" is allowed`)
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

	createArtifactRequest := connectcustompluginv1.ConnectV1CustomConnectorPlugin{
		DisplayName:   connectcustompluginv1.PtrString(displayName),
		Description:   connectcustompluginv1.PtrString(description),
		ConnectorType: connectcustompluginv1.PtrString("flink-udf"),
		UploadSource: &connectcustompluginv1.ConnectV1CustomConnectorPluginUploadSourceOneOf{
			ConnectV1UploadSourcePresignedUrl: connectcustompluginv1.NewConnectV1UploadSourcePresignedUrl("PRESIGNED_URL_LOCATION", resp.GetUploadId()),
		},
	}

	pluginResp, err := c.V2Client.CreateCustomPlugin(createArtifactRequest)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&pluginCreateOut{
		Id:             pluginResp.GetId(),
		Name:           pluginResp.GetDisplayName(),
		ConnectorClass: pluginResp.GetConnectorClass(),
		ContentFormat:  pluginResp.GetContentFormat(),
		Scope:          "org",
	})
	return table.Print()
}
