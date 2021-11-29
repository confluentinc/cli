package connect

import (
	"context"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newResumeCommand() *cobra.Command {
	return &cobra.Command{
		Use:         "resume <id>",
		Short:       "Resume a connector.",
		Args:        cobra.ExactArgs(1),
		RunE:        pcmd.NewCLIRunE(c.resume),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Resume a connector in the current or specified Kafka cluster context.",
				Code: "confluent connect resume --config config.json",
			},
			examples.Example{
				Code: "confluent connect resume --config config.json --cluster lkc-123456",
			},
		),
	}
}

func (c *command) resume(cmd *cobra.Command, args []string) error {
	kafkaCluster, err := c.Context.GetKafkaClusterForCommand(cmd)
	if err != nil {
		return err
	}

	connector := &schedv1.Connector{
		AccountId:      c.EnvironmentId(),
		KafkaClusterId: kafkaCluster.ID,
		Id:             args[0],
	}

	connectorExpansion, err := c.Client.Connect.GetExpansionById(context.Background(), connector)
	if err != nil {
		return err
	}

	connector = &schedv1.Connector{
		Name:           connectorExpansion.Info.Name,
		AccountId:      c.EnvironmentId(),
		KafkaClusterId: kafkaCluster.ID,
	}

	if err := c.Client.Connect.Resume(context.Background(), connector); err != nil {
		return err
	}

	utils.Printf(cmd, errors.ResumedConnectorMsg, args[0])
	return nil
}
