package kafka

import (
	"strconv"
	"strings"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
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
	partitionCommand.AddCommand(partitionCommand.newListCommand())
	partitionCommand.AddCommand(partitionCommand.newDescribeCommand())
	partitionCommand.AddCommand(partitionCommand.newGetReassignmentsCommand())
	return partitionCommand.Command
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
