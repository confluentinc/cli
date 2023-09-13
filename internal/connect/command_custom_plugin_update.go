package connect

import (
	"github.com/spf13/cobra"

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
	cmd.Flags().String("sensitive-properties", "", "Sensitive properties of custom plugin.")
	pcmd.AddContextFlag(cmd, c.CLICommand)

	return cmd
}

func (c *customPluginCommand) update(cmd *cobra.Command, args []string) error {
	if err := errors.CheckNoUpdate(cmd.Flags(), "name", "description", "documentation-link", "sensitive-properties"); err != nil {
		return err
	}
	id := args[0]
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}
	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}
	documentationLink, err := cmd.Flags().GetString("documentation-link")
	if err != nil {
		return err
	}
	sensitivePropertiesString, err := cmd.Flags().GetString("sensitive-properties")
	if err != nil {
		return err
	}

	customPlugin, err := c.V2Client.UpdateCustomPlugin(id, name, description, documentationLink, sensitivePropertiesString)
	if err != nil {
		return err
	}

	output.Printf(errors.UpdatedResourceMsg, resource.CustomConnectorPlugin, args[0])

	table := output.NewTable(cmd)
	table.Add(&pluginCreateOut{
		Id:   customPlugin.GetId(),
		Name: customPlugin.GetDisplayName(),
	})
	return table.Print()
}
