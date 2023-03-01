package kafka

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *quotaCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a Kafka client quota.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.delete,
	}

	pcmd.AddForceFlag(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *quotaCommand) delete(cmd *cobra.Command, args []string) error {
	if _, err := c.V2Client.DescribeKafkaQuota(args[0]); err != nil {
		return err
	}

	promptMsg := fmt.Sprintf(errors.DeleteResourceConfirmMsg, resource.ClientQuota, args[0], args[0])
	if _, err := form.ConfirmDeletion(cmd, promptMsg, args[0]); err != nil {
		return err
	}

	if err := c.V2Client.DeleteKafkaQuota(args[0]); err != nil {
		return err
	}

	utils.Printf(errors.DeletedResourceMsg, resource.ClientQuota, args[0])
	return nil
}
