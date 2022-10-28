package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *mirrorCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <destination-topic-name>",
		Short: "Describe a mirror topic.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.describe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe mirror topic "my-topic" under the link "my-link":`,
				Code: "confluent kafka mirror describe my-topic --link my-link",
			},
		),
	}

	cmd.Flags().String(linkFlagName, "", "Cluster link name.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired(linkFlagName)

	return cmd
}

func (c *mirrorCommand) describe(cmd *cobra.Command, args []string) error {
	mirrorTopicName := args[0]

	linkName, err := cmd.Flags().GetString(linkFlagName)
	if err != nil {
		return err
	}

	kafkaREST, err := c.GetKafkaREST()
	if kafkaREST == nil {
		if err != nil {
			return err
		}
		return errors.New(errors.RestProxyNotAvailableMsg)
	}

	lkc, err := getKafkaClusterLkcId(c.AuthenticatedStateFlagCommand)
	if err != nil {
		return err
	}

	mirror, httpResp, err := kafkaREST.Client.ClusterLinkingV3Api.ReadKafkaMirrorTopic(kafkaREST.Context, lkc, linkName, mirrorTopicName)
	if err != nil {
		return kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
	}

	list := output.NewList(cmd)
	for _, partitionLag := range mirror.MirrorLags {
		list.Add(&mirrorOut{
			LinkName:              mirror.LinkName,
			MirrorTopicName:       mirror.MirrorTopicName,
			SourceTopicName:       mirror.SourceTopicName,
			MirrorStatus:          string(mirror.MirrorStatus),
			StatusTimeMs:          mirror.StateTimeMs,
			Partition:             partitionLag.Partition,
			PartitionMirrorLag:    partitionLag.Lag,
			LastSourceFetchOffset: partitionLag.LastSourceFetchOffset,
		})
	}
	list.Filter([]string{"LinkName", "MirrorTopicName", "Partition", "PartitionMirrorLag", "SourceTopicName", "MirrorStatus", "StatusTimeMs", "LastSourceFetchOffset"})
	return list.Print()
}
