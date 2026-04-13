package kafka

import (
	"github.com/spf13/cobra"
)

type streamsGroupSubtopologyOut struct {
	Kind          string   `human:"Kind" serialized:"kind"`
	ClusterId     string   `human:"Cluster Id" serialized:"cluster_id"`
	GroupId       string   `human:"Group Id" serialized:"group_id"`
	SubtopologyId string   `human:"Subtopology Id" serialized:"subtopology_id"`
	SourceTopics  []string `human:"Source Topics" serialized:"source_topics"`
}

func (c *streamsGroupCommand) newStreamsGroupSubtopologyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subtopology",
		Short: "Manage Kafka stream group subtopologies.",
	}

	cmd.AddCommand(c.newStreamsGroupSubtopologyDescribeCommand())
	cmd.AddCommand(c.newStreamsGroupSubtopologyListCommand())

	return cmd
}
