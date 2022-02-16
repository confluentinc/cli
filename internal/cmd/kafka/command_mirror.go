package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

const (
	replicationFactorFlagName = "replication-factor"
	mirrorStatusFlagName      = "mirror-status"
	linkFlagName              = "link"
)

var (
	listMirrorFields               = []string{"LinkName", "MirrorTopicName", "NumPartition", "MaxPerPartitionMirrorLag", "SourceTopicName", "MirrorStatus", "StatusTimeMs"}
	structuredListMirrorFields     = camelToSnake(listMirrorFields)
	humanListMirrorFields          = camelToSpaced(listMirrorFields)
	describeMirrorFields           = []string{"LinkName", "MirrorTopicName", "Partition", "PartitionMirrorLag", "SourceTopicName", "MirrorStatus", "StatusTimeMs", "LastSourceFetchOffset"}
	structuredDescribeMirrorFields = camelToSnake(describeMirrorFields)
	humanDescribeMirrorFields      = camelToSpaced(describeMirrorFields)
	alterMirrorFields              = []string{"MirrorTopicName", "Partition", "PartitionMirrorLag", "ErrorMessage", "ErrorCode", "LastSourceFetchOffset"}
	structuredAlterMirrorFields    = camelToSnake(alterMirrorFields)
	humanAlterMirrorFields         = camelToSpaced(alterMirrorFields)
)

type listMirrorWrite struct {
	LinkName                 string
	MirrorTopicName          string
	SourceTopicName          string
	MirrorStatus             string
	StatusTimeMs             int32
	NumPartition             int32
	MaxPerPartitionMirrorLag int32
}

type describeMirrorWrite struct {
	LinkName              string
	MirrorTopicName       string
	SourceTopicName       string
	MirrorStatus          string
	StatusTimeMs          int32
	Partition             int32
	PartitionMirrorLag    int32
	LastSourceFetchOffset int64
}

type alterMirrorWrite struct {
	MirrorTopicName       string
	Partition             int32
	ErrorMessage          string
	ErrorCode             string
	PartitionMirrorLag    int64
	LastSourceFetchOffset int64
}

type mirrorCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
}

func newMirrorCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "mirror",
		Short:       "Manages cluster linking mirror topics.",
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
