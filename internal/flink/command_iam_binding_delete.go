package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/deletion"
	"github.com/confluentinc/cli/v3/pkg/resource"
	"github.com/confluentinc/cli/v3/pkg/types"
)

func (c *command) newIamBindingDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id-1> [id-2] ... [id-n]",
		Short: "Delete one or more Flink IAM bindings.",
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

	iamBindingsSet, err := c.getIamBindingsSet(environmentId)
	if err != nil {
		return err
	}

	existenceFunc := func(id string) bool {
		return iamBindingsSet.Contains(id)
	}

	if err := deletion.ValidateAndConfirmDeletion(cmd, args, existenceFunc, resource.FlinkIamBinding, args[0]); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		return c.V2Client.DeleteFlinkIAMBinding(id, environmentId)
	}

	_, err = deletion.Delete(args, deleteFunc, resource.FlinkIamBinding)
	return err
}

func (c *command) getIamBindingsSet(environmentId string) (types.Set[string], error) {
	iamBindings, err := c.V2Client.ListFlinkIAMBindings(environmentId, "", "", "")
	if err != nil {
		return nil, err
	}
	iamBindingsSet := types.NewSet[string]()
	for _, iamBinding := range iamBindings {
		iamBindingsSet.Add(iamBinding.GetId())
	}

	return iamBindingsSet, nil
}
