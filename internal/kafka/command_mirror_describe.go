package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
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
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired(linkFlagName))

	return cmd
}

func (c *mirrorCommand) describe(cmd *cobra.Command, args []string) error {
	mirrorTopicName := args[0]

	linkName, err := cmd.Flags().GetString(linkFlagName)
	if err != nil {
		return err
	}

	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return err
	}

	mirror, err := kafkaREST.CloudClient.ReadKafkaMirrorTopic(linkName, mirrorTopicName)
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
			StatusTimeMs:          mirror.GetStateTimeMs(),
			Partition:             partitionLag.GetPartition(),
			PartitionMirrorLag:    partitionLag.GetLag(),
			LastSourceFetchOffset: partitionLag.GetLastSourceFetchOffset(),
		})
	}
	list.Filter([]string{"LinkName", "MirrorTopicName", "Partition", "PartitionMirrorLag", "SourceTopicName", "MirrorStatus", "StatusTimeMs", "LastSourceFetchOffset"})
	return list.Print()
}
