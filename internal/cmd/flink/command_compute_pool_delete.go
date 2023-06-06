package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *command) newComputePoolDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-n]",
		Short:             "Delete Flink compute pools.",
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

	if err := c.confirmDeletionComputePool(cmd, environmentId, args); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		if err := c.V2Client.DeleteFlinkComputePool(id, environmentId); err != nil {
			return err
		}
		return nil
	}

	deleted, err := resource.Delete(args, deleteFunc, c.postProcess)
	resource.PrintDeleteSuccessMsg(deleted, resource.FlinkComputePool)

	return err
}

func (c *command) confirmDeletionComputePool(cmd *cobra.Command, environmentId string, args []string) error {
	var displayName string
	describeFunc := func(id string) error {
		computePool, err := c.V2Client.DescribeFlinkComputePool(id, environmentId)
		if err == nil && id == args[0] {
			displayName = computePool.Spec.GetDisplayName()
		}
		return err
	}

	if err := resource.ValidateArgs(pcmd.FullParentName(cmd), args, resource.FlinkComputePool, describeFunc); err != nil {
		return err
	}

	if len(args) == 1 {
		if err := form.ConfirmDeletionWithString(cmd, form.DefaultPromptString(resource.FlinkComputePool, args[0], displayName), displayName); err != nil {
			return err
		}
	} else {
		if ok, err := form.ConfirmDeletionYesNo(cmd, form.DefaultYesNoPromptString(resource.FlinkComputePool, args)); err != nil || !ok {
			return err
		}
	}

	return nil
}

func (c *command) postProcess(id string) error {
	if id == c.Context.GetCurrentFlinkComputePool() {
		if err := c.Context.SetCurrentFlinkComputePool(""); err != nil {
			return err
		}
		if err := c.Config.Save(); err != nil {
			return err
		}
	}

	return nil
}
