package ccpm

import (
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/spf13/cobra"
)

func (c *pluginCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <id>",
		Short: "Describe a Custom Connect Plugin.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.describe,
	}

	cmd.Flags().String("environment", "", "Environment ID.")
	cobra.CheckErr(cmd.MarkFlagRequired("environment"))
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *pluginCommand) describe(cmd *cobra.Command, args []string) error {
	pluginId := args[0]

	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	// Use V2Client to call CCPM API
	plugin, err := c.V2Client.DescribeCCPMPlugin(pluginId, environment)
	if err != nil {
		return err
	}
	return printCustomConnectPluginTable(cmd, plugin)
}
