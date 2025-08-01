package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *mirrorCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <destination-topic-name>",
		Short:             "Describe a mirror topic.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe mirror topic "my-topic" under the link "my-link":`,
				Code: "confluent kafka mirror describe my-topic --link my-link",
			},
		),
	}

	pcmd.AddLinkFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEndpointFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("link"))

	return cmd
}

func (c *mirrorCommand) describe(cmd *cobra.Command, args []string) error {
	mirrorTopicName := args[0]

	link, err := cmd.Flags().GetString("link")
	if err != nil {
		return err
	}

	kafkaREST, err := c.GetKafkaREST(cmd)
	if err != nil {
		return err
	}

	mirror, err := kafkaREST.CloudClient.ReadKafkaMirrorTopic(link, mirrorTopicName)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, partitionLag := range mirror.GetMirrorLags().Items {
		list.Add(&mirrorOut{
			LinkName:              mirror.GetLinkName(),
			MirrorTopicName:       mirror.GetMirrorTopicName(),
			SourceTopicName:       mirror.GetSourceTopicName(),
			MirrorStatus:          string(mirror.GetMirrorStatus()),
			MirrorTopicError:      mirror.GetMirrorTopicError(),
			StatusTimeMs:          mirror.GetStateTimeMs(),
			Partition:             partitionLag.GetPartition(),
			PartitionMirrorLag:    partitionLag.GetLag(),
			LastSourceFetchOffset: partitionLag.GetLastSourceFetchOffset(),
		})
	}
	list.Filter([]string{"LinkName", "MirrorTopicName", "Partition", "PartitionMirrorLag", "SourceTopicName", "MirrorStatus", "MirrorTopicError", "StatusTimeMs", "LastSourceFetchOffset"})
	return list.Print()
}
