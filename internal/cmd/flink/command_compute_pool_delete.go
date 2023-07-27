package flink

import (
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *command) newComputePoolDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-n]",
		Short:             "Delete one or more Flink compute pools.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validComputePoolArgsMultiple),
		RunE:              c.computePoolDelete,
	}

	pcmd.AddForceFlag(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)

	return cmd
}

func (c *command) computePoolDelete(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	if confirm, err := c.confirmDeletionComputePool(cmd, environmentId, args); err != nil {
		return err
	} else if !confirm {
		return nil
	}

	deleteFunc := func(id string) error {
		if err := c.V2Client.DeleteFlinkComputePool(id, environmentId); err != nil {
			return err
		}
		return nil
	}

	deletedIDs, err := resource.Delete(args, deleteFunc)
	resource.PrintDeleteSuccessMsg(deletedIDs, resource.FlinkComputePool)

	if err2 := c.removePoolFromConfigIfCurrent(deletedIDs); err2 != nil {
		err = multierror.Append(err, err2)
	}

	return err
}

func (c *command) confirmDeletionComputePool(cmd *cobra.Command, environmentId string, args []string) (bool, error) {
	var displayName string
	describeFunc := func(id string) error {
		computePool, err := c.V2Client.DescribeFlinkComputePool(id, environmentId)
		if err != nil {
			return err
		}
		if id == args[0] {
			displayName = computePool.Spec.GetDisplayName()
		}

		return nil
	}

	if err := resource.ValidateArgs(pcmd.FullParentName(cmd), args, resource.FlinkComputePool, describeFunc); err != nil {
		return false, err
	}

	if len(args) > 1 {
		return form.ConfirmDeletionYesNo(cmd, form.DefaultYesNoPromptString(resource.FlinkComputePool, args))
	}

	if err := form.ConfirmDeletionWithString(cmd, form.DefaultPromptString(resource.FlinkComputePool, args[0], displayName), displayName); err != nil {
		return false, err
	}

	return true, nil
}

func (c *command) removePoolFromConfigIfCurrent(deletedIDs []string) error {
	errs := &multierror.Error{ErrorFormat: errors.CustomMultierrorList}
	for _, id := range deletedIDs {
		if id == c.Context.GetCurrentFlinkComputePool() {
			if err := c.Context.SetCurrentFlinkComputePool(""); err != nil {
				errs = multierror.Append(errs, err)
			}
			if err := c.Config.Save(); err != nil {
				errs = multierror.Append(errs, err)
			}
		}
	}

	return errs.ErrorOrNil()
}
