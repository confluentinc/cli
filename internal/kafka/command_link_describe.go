package kafka

import (
	"github.com/spf13/cobra"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type linkOut struct {
	Name               string            `human:"Name" serialized:"link_name"`
	TopicName          string            `human:"Topic Name" serialized:"topic_name"`
	SourceCluster      string            `human:"Source Cluster" serialized:"source_cluster,omitempty"`
	DestinationCluster string            `human:"Destination Cluster" serialized:"destination_cluster,omitempty"`
	RemoteCluster      string            `human:"Remote Cluster" serialized:"remote_cluster,omitempty"`
	State              string            `human:"State" serialized:"state"`
	Error              string            `human:"Error,omitempty" serialized:"error,omitempty"`
	ErrorMessage       string            `human:"Error Message,omitempty" serialized:"error_message,omitempty"`
	CategoryCounts     []linkCategoryOut `human:"Mirror Partition States,omitempty" serialized:"category_counts,omitempty"`
}

type linkCategoryOut struct {
	StateCategory string `serialized:"state_category"`
	Count         int32  `serialized:"count"`
}

func (c *linkCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <link>",
		Short:             "Describe a cluster link.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describe,
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *linkCommand) describe(cmd *cobra.Command, args []string) error {
	linkName := args[0]

	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return err
	}

	link, err := kafkaREST.CloudClient.GetKafkaLink(linkName)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(newDescribeLink(link, ""))
	table.Filter(getDescribeFields(false))
	return table.Print()
}

func newDescribeLink(link kafkarestv3.ListLinksResponseData, topic string) *linkOut {
	var linkError string
	if link.GetLinkError() != "NO_ERROR" {
		linkError = link.GetLinkError()
	}
	linkCategories := link.GetCategoryCounts()
	categories := make([]linkCategoryOut, len(linkCategories))
	for i, category := range linkCategories {
		categories[i] = linkCategoryOut{
			StateCategory: category.StateCategory,
			Count:         category.Count,
		}
	}
	return &linkOut{
		Name:               link.GetLinkName(),
		TopicName:          topic,
		SourceCluster:      link.GetSourceClusterId(),
		DestinationCluster: link.GetDestinationClusterId(),
		RemoteCluster:      link.GetRemoteClusterId(),
		State:              link.GetLinkState(),
		Error:              linkError,
		ErrorMessage:       link.GetLinkErrorMessage(),
		CategoryCounts:     categories,
	}
}

func getDescribeFields(includeTopics bool) []string {
	x := []string{"Name"}

	if includeTopics {
		x = append(x, "TopicName")
	}

	return append(x, "SourceCluster", "DestinationCluster", "RemoteCluster", "State", "Error", "ErrorMessage", "CategoryCounts")
}
