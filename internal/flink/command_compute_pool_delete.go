package flink

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/resource"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

func (c *command) newComputePoolDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-n]",
		Short:             "Delete one or more Flink compute pools.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validComputePoolArgsMultiple),
		Annotations:       map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
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

	existenceFunc := func(id string) bool {
		_, err := c.V2Client.DescribeFlinkComputePool(id, environmentId)
		return err == nil
	}

	if err := validateAndConfirmComputePoolDeletion(cmd, args, existenceFunc, resource.FlinkComputePool); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		return c.V2Client.DeleteFlinkComputePool(id, environmentId)
	}

	deletedIds, err := deletion.Delete(cmd, args, deleteFunc, resource.FlinkComputePool)

	errs := multierror.Append(err, c.removePoolFromConfigIfCurrent(deletedIds))

	return errs.ErrorOrNil()
}

func confirmComputePoolDeletionString(idList []string) string {
	retentionTimeMsg := "All statements leveraging the compute pool will be STOPPED immediately and be available for 30 days in the statement list history.\n" +
		"After that, they will be permanently deleted."

	if len(idList) == 1 {
		return fmt.Sprintf("Are you sure you want to delete %s \"%s\"? "+retentionTimeMsg, resource.FlinkComputePool, idList[0])
	} else {
		return fmt.Sprintf("Are you sure you want to delete %ss %s? "+retentionTimeMsg, resource.FlinkComputePool, utils.ArrayToCommaDelimitedString(idList, "and"))
	}
}

func validateAndConfirmComputePoolDeletion(cmd *cobra.Command, args []string, checkExistence func(string) bool, resourceType string) error {
	if err := resource.ValidatePrefixes(resourceType, args); err != nil {
		return err
	}

	if err := resource.ValidateArgs(cmd, args, resourceType, checkExistence); err != nil {
		return err
	}

	return deletion.ConfirmPrompt(cmd, confirmComputePoolDeletionString(args))
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
