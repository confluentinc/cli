package kafka

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/spf13/cobra"
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
	req := c.V2Client.KafkaQuotasClient.ClientQuotasKafkaQuotasV1Api.DeleteKafkaQuotasV1ClientQuota(c.quotaContext(), args[0])
	_, err := req.Execute()
	if err != nil {
		return err
	}
	utils.Printf(cmd, errors.DeletedClientQuotaMessage, args[0])
	return nil
}
