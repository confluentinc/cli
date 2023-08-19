package flink

import (
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/deletion"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/resource"
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

	computePool, err := c.V2Client.DescribeFlinkComputePool(args[0], environmentId)
	if err != nil {
		return resource.ResourcesNotFoundError(cmd, resource.FlinkComputePool, args[0])
	}

	existenceFunc := func(id string) bool {
		_, err := c.V2Client.DescribeFlinkComputePool(id, environmentId)
		return err == nil
	}

	if err := deletion.ValidateAndConfirmDeletion(cmd, args, existenceFunc, resource.FlinkComputePool, computePool.Spec.GetDisplayName()); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		return c.V2Client.DeleteFlinkComputePool(id, environmentId)
	}

	deletedIDs, err := deletion.Delete(args, deleteFunc, resource.FlinkComputePool)

	errs := multierror.Append(err, c.removePoolFromConfigIfCurrent(deletedIDs))

	return errs.ErrorOrNil()
}

func (c *command) removePoolFromConfigIfCurrent(deletedIDs []string) error {
	errs := &multierror.Error{ErrorFormat: errors.CustomMultierrorList}
	for _, id := range deletedIDs {
		if id == c.Context.GetCurrentFlinkComputePool() {
			errs = multierror.Append(errs, c.Context.SetCurrentFlinkComputePool(""), c.Config.Save())
		}
	}

	return errs.ErrorOrNil()
}
