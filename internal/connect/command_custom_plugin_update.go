package connect

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/spf13/cobra"
)

func (c *customPluginCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "update <id>",
		Short:       "Update a custom connector plugin configuration.",
		Args:        cobra.ExactArgs(1),
		RunE:        c.update,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}
	cmd.Flags().String("name", "", "name of plugin")
	cmd.Flags().String("description", "", "description of plugin")
	cmd.Flags().String("documentation-link", "", "document link of plugin")
	cmd.Flags().String("sensitive-properties", "", "sensitive properties config of custom plugin")

	pcmd.AddContextFlag(cmd, c.CLICommand)
	return cmd
}

func (c *customPluginCommand) update(cmd *cobra.Command, args []string) error {
	id := args[0]
	name, err := cmd.Flags().GetString("name")
	description, err := cmd.Flags().GetString("description")
	documentationLink, err := cmd.Flags().GetString("documentation-link")
	sensitivePropertiesString, err := cmd.Flags().GetString("sensitive-properties")

	pluginResp, err := c.V2Client.UpdateCustomPlugin(id, name, description, documentationLink, sensitivePropertiesString)
	if err != nil {
		return err
	}

	output.Printf(errors.UpdatedResourceMsg, resource.CustomConnectorPlugin, args[0])

	table := output.NewTable(cmd)
	table.Add(&pluginCreateOut{
		Id:   pluginResp.GetId(),
		Name: pluginResp.GetDisplayName(),
	})
	return table.Print()
}
