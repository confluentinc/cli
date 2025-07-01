package ccpm

import (
	"github.com/spf13/cobra"

	ccpmv1 "github.com/confluentinc/ccloud-sdk-go-v2/ccpm/v1"
	"github.com/confluentinc/cli/v4/pkg/output"
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

	return cmd
}

func (c *pluginCommand) update(cmd *cobra.Command, args []string) error {
	pluginId := args[0]

	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	// Create update request
	update := ccpmv1.CcpmV1CustomConnectPluginUpdate{
		Spec: &ccpmv1.CcpmV1CustomConnectPluginSpecUpdate{},
	}

	if name != "" {
		update.Spec.DisplayName = &name
	}
	if description != "" {
		update.Spec.Description = &description
	}

	// Use V2Client to call CCPM API
	plugin, err := c.V2Client.UpdateCCPMPlugin(pluginId, update)
	if err != nil {
		return err
	}

	output.Printf(c.Config.EnableColor, "Updated Custom Connect Plugin \"%s\".\n", plugin.GetId())

	return nil
}
