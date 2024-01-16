package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
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
	cmd.AddCommand(c.newStateTransitionErrorsCommand())
	cmd.AddCommand(c.newFailoverCommand())
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newPauseCommand())
	cmd.AddCommand(c.newPromoteCommand())
	cmd.AddCommand(c.newResumeCommand())
	cmd.AddCommand(c.newReverseAndPauseMirrorCommand())
	cmd.AddCommand(c.newReverseAndStartMirrorCommand())

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
	link, err := cmd.Flags().GetString("link")
	if err != nil || link == "" {
		return nil
	}

	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return nil
	}

	mirrors, err := kafkaREST.CloudClient.ListKafkaMirrorTopicsUnderLink(link, nil)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(mirrors))
	for i, mirror := range mirrors {
		suggestions[i] = mirror.GetMirrorTopicName()
	}
	return suggestions
}
