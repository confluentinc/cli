package kafka

import (
	"github.com/spf13/cobra"

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
	*pcmd.AuthenticatedStateFlagCommand
}

func newMirrorCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "mirror",
		Short:       "Manage cluster linking mirror topics.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &mirrorCommand{pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)}

	c.AddCommand(c.newCreateCommand())
	c.AddCommand(c.newDescribeCommand())
	c.AddCommand(c.newFailoverCommand())
	c.AddCommand(c.newListCommand())
	c.AddCommand(c.newPauseCommand())
	c.AddCommand(c.newPromoteCommand())
	c.AddCommand(c.newResumeCommand())

	return c.Command
}
