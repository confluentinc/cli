package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type replicaCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
}

var (
	replicaStatusListFields = []string{"ClusterId", "BrokerId", "TopicName", "PartitionId", "IsLeader", "IsObserver", "IsIsrEligible", "IsInIsr", "IsCaughtUp", "LogStartOffset", "LogEndOffset", "LastCaughtUpTimeMs", "LastFetchTimeMs", "LinkName"}
	replicaHumanFields      = []string{"Cluster ID", "Broker ID", "Topic Name", "Partition ID", "Leader", "Observer", "Isr Eligible", "In Isr", "Caught Up", "Log Start Offset", "Log End Offset", "Last Caught Up Time Ms", "Last Fetch Time Ms", "Link Name"}
)

func newReplicaCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "replica",
		Short:       "Manage Kafka replicas.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
	}

	c := &replicaCommand{pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)}
	c.PersistentPreRunE = prerunner.InitializeOnPremKafkaRest(c.AuthenticatedCLICommand)

	cmd.AddCommand(c.newListCommand())

	return cmd
}
