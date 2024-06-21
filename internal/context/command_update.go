package context

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/kafka"
)

func (c *command) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update [context]",
		Short:             "Update a context field.",
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.update,
	}

	cmd.Flags().String("kafka-cluster", "", "Set the active Kafka cluster for the context.")
	pcmd.AddOutputFlag(cmd)

	cmd.MarkFlagRequired("kafka-cluster")

	return cmd
}

func (c *command) update(cmd *cobra.Command, args []string) error {
	ctx, err := c.context(args)
	if err != nil {
		return err
	}

	kafkaCluster, err := cmd.Flags().GetString("kafka-cluster")
	if err != nil {
		return err
	}

	if kafkaCluster != "" {
		if _, err := kafka.FindCluster(nil, ctx, kafkaCluster); err != nil {
			return err
		}

		ctx.KafkaClusterContext.SetActiveKafkaCluster(kafkaCluster)
		if err := ctx.Save(); err != nil {
			return err
		}
	}

	return describeContext(cmd, ctx)
}
