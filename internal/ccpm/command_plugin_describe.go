package ccpm

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *pluginCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <id>",
		Short: "Describe a Custom Connect Plugin.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.describe,
	}

	return cmd
}

func (c *pluginCommand) describe(cmd *cobra.Command, args []string) error {
	pluginId := args[0]

	// Use V2Client to call CCPM API
	plugin, err := c.V2Client.DescribeCCPMPlugin(pluginId)
	if err != nil {
		return err
	}

	// Display plugin details
	spec, _ := plugin.GetSpecOk()
	env, _ := spec.GetEnvironmentOk()
	output.Printf(c.Config.EnableColor, "ID: %s\n", plugin.GetId())
	output.Printf(c.Config.EnableColor, "Name: %s\n", spec.GetDisplayName())
	output.Printf(c.Config.EnableColor, "Description: %s\n", spec.GetDescription())
	output.Printf(c.Config.EnableColor, "Cloud: %s\n", spec.GetCloud())
	output.Printf(c.Config.EnableColor, "Runtime Language: %s\n", spec.GetRuntimeLanguage())
	output.Printf(c.Config.EnableColor, "Environment: %s\n", env.GetId())

	return nil
}
