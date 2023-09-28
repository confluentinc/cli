package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/deletion"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *groupMappingCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-n]",
		Short:             "Delete one or more group mappings.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgsMultiple),
		RunE:              c.delete,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Delete group mapping "group-123456":`,
				Code: "confluent iam group-mapping delete group-123456",
			},
		),
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *groupMappingCommand) delete(cmd *cobra.Command, args []string) error {
	groupMapping, err := c.V2Client.GetGroupMapping(args[0])
	if err != nil {
		return resource.ResourcesNotFoundError(cmd, resource.SsoGroupMapping, args[0])
	}

	existenceFunc := func(id string) bool {
		_, err := c.V2Client.GetGroupMapping(id)
		return err == nil
	}

	if err := deletion.ValidateAndConfirmDeletion(cmd, args, existenceFunc, resource.SsoGroupMapping, groupMapping.GetDisplayName()); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		return c.V2Client.DeleteGroupMapping(id)
	}

	_, err = deletion.Delete(args, deleteFunc, resource.SsoGroupMapping)
	return err
}
