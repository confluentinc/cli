package iam

import (
	"github.com/spf13/cobra"

	ssov2 "github.com/confluentinc/ccloud-sdk-go-v2/sso/v2"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *groupMappingCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <id>",
		Short:             "Update a group mapping.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.update,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update the description of group mapping "group-123456".`,
				Code: `confluent iam group-mapping update group-123456 --description "updated description"`,
			},
		),
	}

	cmd.Flags().String("name", "", "Name of the group mapping.")
	cmd.Flags().String("description", "", "Description of the group mapping.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddFilterFlag(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *groupMappingCommand) update(cmd *cobra.Command, args []string) error {
	if err := errors.CheckNoUpdate(cmd.Flags(), "description", "name", "filter"); err != nil {
		return err
	}

	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	filter, err := cmd.Flags().GetString("filter")
	if err != nil {
		return err
	}

	update := ssov2.IamV2SsoGroupMapping{Id: ssov2.PtrString(args[0])}
	if name != "" {
		update.DisplayName = ssov2.PtrString(name)
	}
	if description != "" {
		update.Description = ssov2.PtrString(description)
	}
	if filter != "" {
		update.Filter = ssov2.PtrString(filter)
	}

	groupMapping, err := c.V2Client.UpdateGroupMapping(update)
	if err != nil {
		return err
	}
	return printGroupMapping(cmd, groupMapping)
}
