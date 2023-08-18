package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/form"
	"github.com/confluentinc/cli/v3/pkg/resource"
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
		return c.V2Client.DeleteKafkaQuota(id)
	}

	_, err := resource.Delete(args, deleteFunc, resource.ClientQuota)
	return err
}

func (c *quotaCommand) confirmDeletion(cmd *cobra.Command, args []string) (bool, error) {
	var displayName string
	existenceFunc := func(id string) bool {
		quota, err := c.V2Client.DescribeKafkaQuota(id)
		if err != nil {
			return false
		}
		if id == args[0] {
			displayName = quota.Spec.GetDisplayName()
		}

		return true
	}

	if err := resource.ValidateArgs(pcmd.FullParentName(cmd), args, resource.ClientQuota, existenceFunc); err != nil {
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
