package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *quotaCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-n]",
		Short:             "Delete one or more Kafka client quotas.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgsMultiple),
		RunE:              c.delete,
	}

	pcmd.AddForceFlag(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *quotaCommand) delete(cmd *cobra.Command, args []string) error {
	if confirm, err := c.confirmDeletion(cmd, args); err != nil {
		return err
	} else if !confirm {
		return nil
	}

	deleteFunc := func(id string) error {
		if err := c.V2Client.DeleteKafkaQuota(id); err != nil {
			return err
		}
		return nil
	}

	deleted, err := resource.Delete(args, deleteFunc)
	resource.PrintDeleteSuccessMsg(deleted, resource.ClientQuota)

	return err
}

func (c *quotaCommand) confirmDeletion(cmd *cobra.Command, args []string) (bool, error) {
	var displayName string
	describeFunc := func(id string) error {
		quota, err := c.V2Client.DescribeKafkaQuota(id)
		if err != nil {
			return err
		}
		if id == args[0] {
			displayName = quota.Spec.GetDisplayName()
		}

		return nil
	}

	if err := resource.ValidateArgs(pcmd.FullParentName(cmd), args, resource.ClientQuota, describeFunc); err != nil {
		return false, err
	}

	if len(args) > 1 {
		return form.ConfirmDeletionYesNo(cmd, form.DefaultYesNoPromptString(resource.ClientQuota, args))
	}

	if err := form.ConfirmDeletionWithString(cmd, form.DefaultPromptString(resource.ClientQuota, args[0], displayName), displayName); err != nil {
		return false, err
	}

	return true, nil
}
