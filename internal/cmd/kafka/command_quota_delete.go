package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *quotaCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a Kafka client quota.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.delete,
	}
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *quotaCommand) delete(cmd *cobra.Command, args []string) error {
	err := c.V2Client.DeleteKafkaQuota(args[0])
	if err != nil {
		return quotaErr(err)
	}
	utils.Printf(cmd, errors.DeletedClientQuotaMessage, args[0])
	return nil
}
