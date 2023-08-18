package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/form"
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
				Text: `Delete group mapping "pool-12345":`,
				Code: "confluent iam group-mapping delete pool-12345",
			},
		),
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *groupMappingCommand) delete(cmd *cobra.Command, args []string) error {
	if confirm, err := c.confirmDeletion(cmd, args); err != nil {
		return err
	} else if !confirm {
		return nil
	}

	if err := c.V2Client.DeleteGroupMapping(args[0]); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		return c.V2Client.DeleteGroupMapping(id)
	}

	_, err := resource.Delete(args, deleteFunc, resource.SsoGroupMapping)
	return err
}

func (c *groupMappingCommand) confirmDeletion(cmd *cobra.Command, args []string) (bool, error) {
	var displayName string
	existenceFunc := func(id string) bool {
		groupMapping, err := c.V2Client.GetGroupMapping(id)
		if err != nil {
			return false
		}
		if id == args[0] {
			displayName = groupMapping.GetDisplayName()
		}

		return true
	}

	if err := resource.ValidateArgs(cmd, args, resource.SsoGroupMapping, existenceFunc); err != nil {
		return false, err
	}

	if len(args) > 1 {
		return form.ConfirmDeletionYesNo(cmd, form.DefaultYesNoPromptString(resource.SsoGroupMapping, args))
	}

	if err := form.ConfirmDeletionWithString(cmd, form.DefaultPromptString(resource.SsoGroupMapping, args[0], displayName), displayName); err != nil {
		return false, err
	}

	return true, nil
}
