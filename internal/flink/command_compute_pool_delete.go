package flink

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/deletion"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/resource"
	"github.com/confluentinc/cli/v3/pkg/utils"
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

func (c *command) validComputePoolArgsMultiple(cmd *cobra.Command, args []string) []string {
	return c.autocompleteComputePools(cmd, args)
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

	if err := validateAndConfirmComputePoolDeletion(cmd, args, existenceFunc, resource.FlinkComputePool, computePool.Spec.GetDisplayName()); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		return c.V2Client.DeleteFlinkComputePool(id, environmentId)
	}

	deletedIds, err := deletion.Delete(args, deleteFunc, resource.FlinkComputePool)

	errs := multierror.Append(err, c.removePoolFromConfigIfCurrent(deletedIds))

	return errs.ErrorOrNil()
}

func confirmDeletionString(name, id string) string {
	deletionString := fmt.Sprintf("Are you sure you want to delete the compute pool \"%s\"?"+
		" All statements leveraging the compute pool will be STOPPED immediately and be available for 30 days in the statement list history.\n"+
		"After that, they will be permanently deleted. \n"+
		"To confirm, type \"%s\". To cancel, press Ctrl-C", id, name)

	return deletionString
}

func confirmMultipleDeletionString(idList []string) string {
	deletionString := fmt.Sprintf("Are you sure you want to delete compute pools %s?"+
		" All statements leveraging the compute pools will be STOPPED immediately and be available for 30 days in the statement list history.\n"+
		"After that, they will be permanently deleted. \n", utils.ArrayToCommaDelimitedString(idList, "and"))

	return deletionString
}

func validateAndConfirmComputePoolDeletion(cmd *cobra.Command, args []string, checkExistence func(string) bool, resourceType, name string) error {
	if err := resource.ValidatePrefixes(resourceType, args); err != nil {
		return err
	}

	if err := resource.ValidateArgs(cmd, args, resourceType, checkExistence); err != nil {
		return err
	}

	if len(args) > 1 {
		return deletion.ConfirmPromptYesOrNo(cmd, confirmMultipleDeletionString(args))
	}

	promptString := confirmDeletionString(name, args[0])
	if err := deletion.ConfirmDeletionWithString(cmd, promptString, name); err != nil {
		return err
	}

	return nil
}

func (c *command) removePoolFromConfigIfCurrent(deletedIds []string) error {
	errs := &multierror.Error{ErrorFormat: errors.CustomMultierrorList}
	for _, id := range deletedIds {
		if id == c.Context.GetCurrentFlinkComputePool() {
			errs = multierror.Append(errs, c.Context.SetCurrentFlinkComputePool(""), c.Config.Save())
		}
	}

	return errs.ErrorOrNil()
}
