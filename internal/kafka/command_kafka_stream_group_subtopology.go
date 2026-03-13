package kafka

import (
	"github.com/spf13/cobra"
)

type streamGroupSubtopologyOut struct {
	Kind          string   `human:"Kind" serialized:"kind"`
	ClusterId     string   `human:"Cluster Id" serialized:"cluster_id"`
	GroupId       string   `human:"Group Id" serialized:"group_id"`
	SubtopologyId string   `human:"Subtopology Id" serialized:"subtopology_id"`
	SourceTopics  []string `human:"Source Topics" serialized:"source_topics"`
}

func (c *consumerCommand) newStreamGroupSubtopologyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stream-group-subtopology",
		Short: "Manage Kafka stream group subtopologies.",
	}

	cmd.AddCommand(c.newStreamGroupSubtopologyDescribeCommand())
	cmd.AddCommand(c.newStreamGroupSubtopologyListCommand())

	return cmd
}
