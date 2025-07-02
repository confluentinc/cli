package ccpm

import (
	"github.com/spf13/cobra"

	ccpmv1 "github.com/confluentinc/ccloud-sdk-go-v2/ccpm/v1"
)

func (c *pluginCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a Custom Connect Plugin.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.update,
	}

	cmd.Flags().String("name", "", "Display name of the Custom Connect Plugin.")
	cmd.Flags().String("description", "", "Description of the Custom Connect Plugin.")
	cmd.Flags().String("environment", "", "Environment ID.")
	cobra.CheckErr(cmd.MarkFlagRequired("environment"))

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
	environment, err := cmd.Flags().GetString("environment")
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
