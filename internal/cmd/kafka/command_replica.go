package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
)

type replicaCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
}

var (
	replicaStatusListFields = []string{"ClusterId", "BrokerId", "TopicName", "PartitionId", "IsLeader", "IsObserver", "IsIsrEligible", "IsInIsr", "IsCaughtUp", "LogStartOffset", "LogEndOffset", "LastCaughtUpTimeMs", "LastFetchTimeMs", "LinkName"}
	replicaHumanFields      = []string{"Cluster ID", "Broker ID", "Topic Name", "Partition ID", "Leader", "Observer", "Isr Eligible", "In Isr", "Caught Up", "Log Start Offset", "Log End Offset", "Last Caught Up Time Ms", "Last Fetch Time Ms", "Link Name"}
)

func newReplicaCommand(prerunner pcmd.PreRunner) *cobra.Command {
	replicaCommand := &replicaCommand{
		AuthenticatedStateFlagCommand: pcmd.NewAuthenticatedStateFlagCommand(
			&cobra.Command{
				Use:         "replica",
				Short:       "Manage Kafka replicas.",
				Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
			}, prerunner),
	}
	replicaCommand.SetPersistentPreRunE(prerunner.InitializeOnPremKafkaRest(replicaCommand.AuthenticatedCLICommand))
	replicaCommand.init()
	return replicaCommand.Command
}

func (replicaCommand *replicaCommand) init() {
	listCmd := &cobra.Command{
		Use:   "list",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(replicaCommand.list),
		Short: "List Kafka replica statuses.",
		Long:  "List partition-replicas statuses filtered by topic and partition via Confluent Kafka REST.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `List the replica statuses for partition 1 of "my_topic".`,
				Code: "confluent kafka replica list --topic my_topic --partition 1",
			},
			examples.Example{
				Text: `List the replicas statuses for topic "my_topic".`,
				Code: "confluent kafka replica list --topic my_topic",
			},
		),
	}
	listCmd.Flags().String("topic", "", "Topic name.")
	listCmd.Flags().Int32("partition", -1, "Partition ID.")
	listCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddOutputFlag(listCmd)
	_ = listCmd.MarkFlagRequired("topic")
	replicaCommand.AddCommand(listCmd)
}
