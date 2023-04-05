package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/deletion"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *quotaCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id-1> [id-2] ... [id-n]",
		Short: "Delete Kafka client quotas.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.delete,
	}

	pcmd.AddForceFlag(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *quotaCommand) delete(cmd *cobra.Command, args []string) error {
	if err := c.validateArgs(cmd, args); err != nil {
		return err
	}

	if len(args) == 1 {
		if err := form.ConfirmDeletionWithString(cmd, resource.ClientQuota, args[0], args[0]); err != nil {
			return err
		}
	} else {
		if ok, err := form.ConfirmDeletionYesNo(cmd, resource.ClientQuota, args); err != nil || !ok {
			return err
		}
	}

	var errs error
	var deleted []string
	for _, id := range args {
		if err := c.V2Client.DeleteKafkaQuota(id); err != nil {
			errs = errors.Join(errs, err)
		} else {
			deleted = append(deleted, id)
		}
	}
	deletion.PrintSuccessfulDeletionMsg(deleted, resource.ClientQuota)

	return errs
}

func (c *quotaCommand) validateArgs(cmd *cobra.Command, args []string) error {
	describeFunc := func(id string) error {
		_, err := c.V2Client.DescribeKafkaQuota(id)
		return err
	}

	return deletion.ValidateArgsForDeletion(cmd, args, resource.ClientQuota, describeFunc)
}
