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
	id := args[0]
	var err error
	var name, description, documentationLink, sensitivePropertiesString string
	if name, err = cmd.Flags().GetString("name"); err != nil {
		return err
	}
	if description, err = cmd.Flags().GetString("description"); err != nil {
		return err
	}
	if documentationLink, err = cmd.Flags().GetString("documentation-link"); err != nil {
		return err
	}
	if sensitivePropertiesString, err = cmd.Flags().GetString("sensitive-properties"); err != nil {
		return err
	}

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
