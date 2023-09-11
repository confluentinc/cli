package connect

import (
	"fmt"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/spf13/cobra"
)

func (c *customPluginCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "delete <id>",
		Short:       "Delete a custom connector plugin.",
		Args:        cobra.ExactArgs(1),
		RunE:        c.delete,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Delete a custom connector plugin",
				Code: "confluent connect custom-plugin delete ccp-123456",
			},
		),
	}

	pcmd.AddForceFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)

	return cmd
}

func (c *customPluginCommand) delete(cmd *cobra.Command, args []string) error {
	pluginId := args[0]
	plugin, err := c.V2Client.DescribeCustomPlugin(pluginId)
	if err != nil {
		return err
	}
	promptMsg := fmt.Sprintf(errors.DeleteResourceConfirmMsg, resource.Connector, pluginId, plugin.GetDisplayName())
	if _, err := form.ConfirmDeletion(cmd, promptMsg, plugin.GetDisplayName()); err != nil {
		return err
	}

	if err := c.V2Client.DeleteCustomPlugin(pluginId); err != nil {
		return err
	}

	output.Printf(errors.DeletedResourceMsg, resource.CustomConnectorPlugin, pluginId)
	return nil
}
