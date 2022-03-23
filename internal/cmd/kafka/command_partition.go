package kafka

import (
	"strconv"
	"strings"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
)

type partitionCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
}

func newPartitionCommand(prerunner pcmd.PreRunner) *cobra.Command {
	partitionCommand := &partitionCommand{
		AuthenticatedStateFlagCommand: pcmd.NewAuthenticatedStateFlagCommand(
			&cobra.Command{
				Use:         "partition",
				Short:       "Manage Kafka partitions.",
				Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
			}, prerunner),
	}
	partitionCommand.SetPersistentPreRunE(prerunner.InitializeOnPremKafkaRest(partitionCommand.AuthenticatedCLICommand))
	partitionCommand.init()
	return partitionCommand.Command
}

func (partitionCmd *partitionCommand) init() {
	listCmd := &cobra.Command{
		Use:   "list",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(partitionCmd.list),
		Short: "List Kafka partitions.",
		Long:  "List the partitions that belong to a specified topic via Confluent Kafka REST.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `List the partitions for "my_topic".`,
				Code: "confluent kafka partition list --topic my_topic",
			},
		),
	}
	listCmd.Flags().String("topic", "", "Topic name to list partitions of.")
	listCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddOutputFlag(listCmd)
	_ = listCmd.MarkFlagRequired("topic")
	partitionCmd.AddCommand(listCmd)

	describeCmd := &cobra.Command{
		Use:   "describe <id>",
		Short: "Describe a Kafka partition.",
		Long:  "Describe a Kafka partition via Confluent Kafka REST.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(partitionCmd.describe),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe partition "1" for "my_topic".`,
				Code: "confluent kafka partition describe 1 --topic my_topic",
			}),
	}
	describeCmd.Flags().String("topic", "", "Topic name to list partitions of.")
	describeCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddOutputFlag(describeCmd)
	_ = describeCmd.MarkFlagRequired("topic")
	partitionCmd.AddCommand(describeCmd)

	reassignmentsCmd := &cobra.Command{
		Use:   "get-reassignments [id]",
		Short: "Get ongoing replica reassignments.",
		Long:  "Get ongoing replica reassignments for a given cluster, topic, or partition via Confluent Kafka REST.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  pcmd.NewCLIRunE(partitionCmd.getReassignments),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Get all replica reassignments for the Kafka cluster.",
				Code: "confluent kafka partition get-reassignments",
			},
			examples.Example{
				Text: `Get replica reassignments for "my_topic".`,
				Code: "confluent kafka partition get-reassignments --topic my_topic",
			},
			examples.Example{
				Text: `Get replica reassignments for partition "1" of "my_topic".`,
				Code: "confluent kafka partition get-reassignments 1 --topic my_topic",
			}),
	}
	reassignmentsCmd.Flags().String("topic", "", "Topic name to search by.")
	reassignmentsCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddOutputFlag(reassignmentsCmd)
	partitionCmd.AddCommand(reassignmentsCmd)
}

func parseLeaderId(leader kafkarestv3.Relationship) int32 {
	index := strings.LastIndex(leader.Related, "/")
	idStr := leader.Related[index+1:]
	leaderId, _ := strconv.ParseInt(idStr, 10, 32)
	return int32(leaderId)
}

func partitionIdFromArg(args []string) (int32, error) {
	partitionIdStr := args[0]
	partitionId, err := strconv.ParseInt(partitionIdStr, 10, 32)
	return int32(partitionId), err
}
