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
	cmd := &cobra.Command{
		Use:         "partition",
		Short:       "Manage Kafka partitions.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
	}

	c := &partitionCommand{pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)}
	c.SetPersistentPreRunE(prerunner.InitializeOnPremKafkaRest(c.AuthenticatedCLICommand))

	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newGetReassignmentsCommand())
	cmd.AddCommand(c.newListCommand())

	return cmd
}

func parseLeaderId(leader kafkarestv3.Relationship) int32 {
	index := strings.LastIndex(leader.Related, "/")
	idStr := leader.Related[index+1:]
	leaderId, _ := strconv.ParseInt(idStr, 10, 32)
	return int32(leaderId)
}

func partitionIdFromArg(args []string) (int32, error) {
	partitionId, err := strconv.ParseInt(args[0], 10, 32)
	return int32(partitionId), err
}
