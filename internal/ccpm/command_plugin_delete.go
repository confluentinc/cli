package ccpm

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *pluginCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a Custom Connect Plugin.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.delete,
	}

	return cmd
}

func (c *pluginCommand) delete(cmd *cobra.Command, args []string) error {
	pluginId := args[0]

	// Use V2Client to call CCPM API
	err := c.V2Client.DeleteCCPMPlugin(pluginId)
	if err != nil {
		return err
	}

	output.Printf(c.Config.EnableColor, "Deleted Custom Connect Plugin \"%s\".\n", pluginId)

	return nil
}
