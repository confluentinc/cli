package connect

import (
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/spf13/cobra"
	"path/filepath"
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
				Text: "Create a custom connector plugin ",
				Code: "confluent connect custom-plugin create plugin_name --config-file config.json",
			},
		),
	}

	cmd.Flags().String("plugin-file", "", "ZIP/JAR plugin file.")
	cmd.Flags().String("description", "", "Description of custom plugin.")
	cmd.Flags().String("documentation-link", "", "Document link of custom plugin.")
	cmd.Flags().String("connector-class", "", "Connector class of custom plugin.")
	cmd.Flags().String("connector-type", "", "Connector type of custom plugin.")
	cmd.Flags().String("sensitive-properties", "", "Sensitive properties of custom plugin.")

	cobra.CheckErr(cmd.MarkFlagRequired("connector-class"))
	cobra.CheckErr(cmd.MarkFlagRequired("connector-type"))

	cobra.CheckErr(cmd.MarkFlagFilename("plugin-file", "zip", "jar"))
	cobra.CheckErr(cmd.MarkFlagRequired("plugin-file"))
	return cmd
}

func (c *customPluginCommand) createCustomPlugin(cmd *cobra.Command, args []string) error {
	displayName := args[0]
	description, err := cmd.Flags().GetString("description")
	documentationLink, err := cmd.Flags().GetString("documentation-link")
	pluginFileName, err := cmd.Flags().GetString("plugin-file")
	connectorClass, err := cmd.Flags().GetString("connector-class")
	connectorType, err := cmd.Flags().GetString("connector-type")
	sensitivePropertiesString, err := cmd.Flags().GetString("sensitive-properties")

	extension := filepath.Ext(pluginFileName)[1:]
	if extension != "zip" && extension != "jar" {
		return errors.Errorf("only ZIP/JAR plugin file is allowed")
	}

	resp, err := c.V2Client.GetPresignedUrl(extension)
	if err != nil {
		return err
	}

	err = uploadFile(resp.GetUploadUrl(), pluginFileName, resp.GetUploadFormData())
	if err != nil {
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
