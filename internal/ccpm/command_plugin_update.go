package ccpm

import (
	"github.com/spf13/cobra"

	ccpmv1 "github.com/confluentinc/ccloud-sdk-go-v2/ccpm/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *pluginCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a custom Connect plugin.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.update,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Update the name and description of a custom Connect plugin.",
				Code: `confluent ccpm plugin update plugin-123456 --name "Updated Plugin Name" --description "Updated description" --environment env-12345`,
			},
			examples.Example{
				Text: "Update only the name of a custom Connect plugin.",
				Code: `confluent ccpm plugin update plugin-123456 --name "New Plugin Name" --environment env-12345`,
			},
		),
	}

	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String("name", "", "Display name of the custom Connect plugin.")
	cmd.Flags().String("description", "", "Description of the custom Connect plugin.")

	return cmd
}

func (c *pluginCommand) update(cmd *cobra.Command, args []string) error {
	pluginId := args[0]

	// Create update request
	updateCustomPluginRequest := ccpmv1.CcpmV1CustomConnectPluginUpdate{
		Spec: &ccpmv1.CcpmV1CustomConnectPluginSpecUpdate{},
	}

	if cmd.Flags().Changed("name") {
		if name, err := cmd.Flags().GetString("name"); err != nil {
			return err
		} else {
			updateCustomPluginRequest.Spec.SetDisplayName(name)
		}
	}
	if cmd.Flags().Changed("description") {
		if description, err := cmd.Flags().GetString("description"); err != nil {
			return err
		} else {
			updateCustomPluginRequest.Spec.SetDescription(description)
		}
	}

	// Get environment ID
	environment, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}
	updateCustomPluginRequest.Spec.Environment = &ccpmv1.EnvScopedObjectReference{
		Id: environment,
	}

	// Use V2Client to call CCPM API
	plugin, err := c.V2Client.UpdateCCPMPlugin(pluginId, updateCustomPluginRequest)
	if err != nil {
		return err
	}
	return printCustomConnectPluginTable(cmd, plugin)
}
