package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/deletion"
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
	quota, err := c.V2Client.DescribeKafkaQuota(args[0])
	if err != nil {
		return resource.ResourcesNotFoundError(cmd, resource.ClientQuota, args[0])
	}

	existenceFunc := func(id string) bool {
		_, err := c.V2Client.DescribeKafkaQuota(id)
		return err == nil
	}

	if err := deletion.ValidateAndConfirmDeletion(cmd, args, existenceFunc, resource.ClientQuota, quota.Spec.GetDisplayName()); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		return c.V2Client.DeleteKafkaQuota(id)
	}

	_, err = deletion.Delete(args, deleteFunc, resource.ClientQuota)
	return err
}
