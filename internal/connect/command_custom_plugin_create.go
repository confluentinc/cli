package connect

import (
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/errors"
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
	cmd.Flags().String("sensitive-properties", "", "Sensitive properties of custom plugin.")

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
	sensitivePropertiesString, err := cmd.Flags().GetString("sensitive-properties")
	if err != nil {
		return err
	}

	extension := filepath.Ext(pluginFileName)
	if extension == "" {
		return errors.Errorf("unable to read plugin file extension")
	}
	extension = extension[1:]
	if extension != "zip" && extension != "jar" {
		return errors.Errorf("only ZIP/JAR plugin files are allowed")
	}

	resp, err := c.V2Client.GetPresignedUrl(extension)
	if err != nil {
		return err
	}

	if err = uploadFile(resp.GetUploadUrl(), pluginFileName, resp.GetUploadFormData()); err != nil {
		return err
	}

	pluginResp, err := c.V2Client.CreateCustomPlugin(displayName, description, documentationLink, connectorClass, connectorType, sensitivePropertiesString, resp.GetUploadId())
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
