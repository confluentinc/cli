package iam

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/form"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *groupMappingCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id>",
		Short:             "Delete a group mapping.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
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
	groupMapping, err := c.V2Client.GetGroupMapping(args[0])
	if err != nil {
		return err
	}

	promptMsg := fmt.Sprintf(errors.DeleteResourceConfirmMsg, resource.SsoGroupMapping, args[0], groupMapping.GetDisplayName())
	if _, err := form.ConfirmDeletion(cmd, promptMsg, groupMapping.GetDisplayName()); err != nil {
		return err
	}

	if err := c.V2Client.DeleteGroupMapping(args[0]); err != nil {
		return err
	}

	output.ErrPrintf(errors.DeletedResourceMsg, resource.SsoGroupMapping, args[0])
	return nil
}
