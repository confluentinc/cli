package kafka

import (
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
)

type partitionCommand struct {
	*pcmd.AuthenticatedCLICommand
}

func newPartitionCommand(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "partition",
		Short:       "Manage Kafka partitions.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLoginOrOnPremLogin},
	}

	c := &partitionCommand{}

	if cfg.IsCloudLogin() {
		c.AuthenticatedCLICommand = pcmd.NewAuthenticatedCLICommand(cmd, prerunner)

		cmd.AddCommand(c.newDescribeCommand())
		cmd.AddCommand(c.newListCommand())
	} else {
		c.AuthenticatedCLICommand = pcmd.NewAuthenticatedWithMDSCLICommand(cmd, prerunner)
		c.PersistentPreRunE = prerunner.InitializeOnPremKafkaRest(c.AuthenticatedCLICommand)

		cmd.AddCommand(c.newDescribeCommandOnPrem())
		cmd.AddCommand(c.newListCommandOnPrem())
		cmd.AddCommand(c.newReassignmentCommand())
	}

	return cmd
}

func parseLeaderId(related string) int32 {
	index := strings.LastIndex(related, "/")
	idStr := related[index+1:]
	leaderId, _ := strconv.ParseInt(idStr, 10, 32)
	return int32(leaderId)
}

func partitionIdFromArg(args []string) (int32, error) {
	partitionId, err := strconv.ParseInt(args[0], 10, 32)
	return int32(partitionId), err
}
