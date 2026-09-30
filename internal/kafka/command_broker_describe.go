package kafka

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/broker"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

func (c *brokerCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <id>",
		Short: "Describe a Kafka broker.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.describe,
	}

	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *brokerCommand) describe(cmd *cobra.Command, args []string) error {
	restClient, restContext, clusterId, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	return broker.Describe(cmd, args, restClient, restContext, clusterId)
}
