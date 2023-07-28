package kafka

import (
	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

const (
	replicationFactorFlagName = "replication-factor"
	mirrorStatusFlagName      = "mirror-status"
	linkFlagName              = "link"
	sourceTopicFlagName       = "source-topic"
)

type mirrorOut struct {
	LinkName                 string `human:"Link Name" serialized:"link_name"`
	MirrorTopicName          string `human:"Mirror Topic Name" serialized:"mirror_topic_name"`
	SourceTopicName          string `human:"Source Topic Name" serialized:"source_topic_name"`
	MirrorStatus             string `human:"Mirror Status" serialized:"mirror_status"`
	StatusTimeMs             int64  `human:"Status Time (ms)" serialized:"status_time_ms"`
	Partition                int32  `human:"Partition" serialized:"partition"`
	NumPartition             int32  `human:"Num Partition" serialized:"num_partition"`
	PartitionMirrorLag       int64  `human:"Partition Mirror Lag" serialized:"partition_mirror_lag"`
	MaxPerPartitionMirrorLag int64  `human:"Max Per Partition Mirror Lag" serialized:"max_per_partition_mirror_lag"`
	LastSourceFetchOffset    int64  `human:"Last Source Fetch Offset" serialized:"last_source_fetch_offset"`
	ErrorMessage             string `human:"Error Message" serialized:"error_message"`
	ErrorCode                string `human:"Error Code" serialized:"error_code"`
}

type mirrorCommand struct {
	*pcmd.AuthenticatedCLICommand
}

func newMirrorCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "mirror",
		Short:       "Manage cluster linking mirror topics.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &mirrorCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newFailoverCommand())
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newPauseCommand())
	cmd.AddCommand(c.newPromoteCommand())
	cmd.AddCommand(c.newResumeCommand())

	return cmd
}

func (c *mirrorCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	return c.validArgsMultiple(cmd, args)
}

func (c *mirrorCommand) validArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompleteMirrorTopics(cmd)
}

func (c *mirrorCommand) autocompleteMirrorTopics(cmd *cobra.Command) []string {
	linkName, err := cmd.Flags().GetString(linkFlagName)
	if err != nil || linkName == "" {
		return nil
	}

	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return nil
	}

	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return nil
	}

	opts := &kafkarestv3.ListKafkaMirrorTopicsUnderLinkOpts{MirrorStatus: optional.EmptyInterface()}
	listMirrorTopicsResponseDataList, _, err := kafkaREST.Client.ClusterLinkingV3Api.ListKafkaMirrorTopicsUnderLink(kafkaREST.Context, cluster.ID, linkName, opts)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(listMirrorTopicsResponseDataList.Data))
	for i, mirrorTopic := range listMirrorTopicsResponseDataList.Data {
		suggestions[i] = mirrorTopic.MirrorTopicName
	}
	return suggestions
}
