package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/types"
)

func (c *command) newIamBindingDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id-1> [id-2] ... [id-n]",
		Short: "Delete Flink IAM bindings.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.iamBindingDelete,
	}

	pcmd.AddForceFlag(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *command) iamBindingDelete(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	if err := c.confirmDeletionIamBinding(cmd, environmentId, args); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		return c.V2Client.DeleteFlinkIAMBinding(id, environmentId)
	}

	deleted, err := resource.Delete(args, deleteFunc, nil)
	resource.PrintDeleteSuccessMsg(deleted, resource.FlinkIamBinding)

	return err
}

func (c *command) confirmDeletionIamBinding(cmd *cobra.Command, environmentId string, args []string) error {
	iamBindings, err := c.V2Client.ListFlinkIAMBindings(environmentId, "", "", "")
	if err != nil {
		return err
	}
	iamBindingsSet := types.NewSet[string]()
	for _, iamBinding := range iamBindings {
		iamBindingsSet.Add(iamBinding.GetId())
	}

	describeFunc := func(id string) error {
		if !iamBindingsSet.Contains(id) {
			return errors.Errorf(`%s "%s" not found`, resource.FlinkIamBinding, id)
		}
		return nil
	}

	if err := resource.ValidateArgs(pcmd.FullParentName(cmd), args, resource.FlinkIamBinding, describeFunc); err != nil {
		return err
	}

	if len(args) == 1 {
		if err := form.ConfirmDeletionWithString(cmd, form.DefaultPromptString(resource.FlinkIamBinding, args[0], args[0]), args[0]); err != nil {
			return err
		}
	} else {
		if ok, err := form.ConfirmDeletionYesNo(cmd, form.DefaultYesNoPromptString(resource.FlinkIamBinding, args)); err != nil || !ok {
			return err
		}
	}

	return nil
}
