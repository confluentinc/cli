package connect

import (
	"github.com/spf13/cobra"

	connectcustompluginv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect-custom-plugin/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *customPluginCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a custom connector plugin configuration.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.update,
	}

	cmd.Flags().String("name", "", "Name of custom plugin.")
	cmd.Flags().String("description", "", "Description of custom plugin.")
	cmd.Flags().String("documentation-link", "", "Document link of custom plugin.")
	cmd.Flags().StringSlice("sensitive-properties", nil, "A comma-separated list of sensitive property names.")
	pcmd.AddContextFlag(cmd, c.CLICommand)

	return cmd
}

func (c *customPluginCommand) update(cmd *cobra.Command, args []string) error {
	if err := errors.CheckNoUpdate(cmd.Flags(), "name", "description", "documentation-link", "sensitive-properties"); err != nil {
		return err
	}

	id := args[0]
	updateCustomPluginRequest := connectcustompluginv1.NewConnectV1CustomConnectorPluginUpdate()

	if cmd.Flags().Changed("name") {
		if name, err := cmd.Flags().GetString("name"); err != nil {
			return err
		} else {
			updateCustomPluginRequest.SetDisplayName(name)
		}
	}
	if cmd.Flags().Changed("description") {
		if description, err := cmd.Flags().GetString("description"); err != nil {
			return err
		} else {
			updateCustomPluginRequest.SetDescription(description)
		}
	}
	if cmd.Flags().Changed("documentation-link") {
		if documentationLink, err := cmd.Flags().GetString("documentation-link"); err != nil {
			return err
		} else {
			updateCustomPluginRequest.SetDocumentationLink(documentationLink)
		}
	}
	if cmd.Flags().Changed("sensitive-properties") {
		if sensitiveProperties, err := cmd.Flags().GetString("sensitive-properties"); err != nil {
			return err
		} else {
			updateCustomPluginRequest.SetDisplayName(sensitiveProperties)
		}
	}

	if _, err := c.V2Client.UpdateCustomPlugin(id, *updateCustomPluginRequest); err != nil {
		return err
	}

	output.Printf(errors.UpdatedResourceMsg, resource.CustomConnectorPlugin, args[0])
	return nil
}
