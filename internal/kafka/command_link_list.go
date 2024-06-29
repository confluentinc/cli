package kafka

import (
	"github.com/spf13/cobra"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
)

const includeTopicsFlagName = "include-topics"

type listOut struct {
	Name                 string `human:"Name" serialized:"link_name"`
	TopicName            string `human:"Topic Name" serialized:"topic_name"`
	SourceClusterId      string `human:"Source Cluster" serialized:"source_cluster_id"`
	DestinationClusterId string `human:"Destination Cluster" serialized:"destination_cluster_id"`
	RemoteClusterId      string `human:"Remote Cluster" serialized:"remote_cluster_id"`
	State                string `human:"State" serialized:"state"`
	Error                string `human:"Error" serialized:"error"`
	ErrorMessage         string `human:"Error Message" serialized:"error_message"`
}

func newLink(link kafkarestv3.ListLinksResponseData, topic string) *listOut {
	var linkError string
	if link.GetLinkError() != "NO_ERROR" {
		linkError = link.GetLinkError()
	}
	return &listOut{
		Name:                 link.GetLinkName(),
		TopicName:            topic,
		SourceClusterId:      link.GetSourceClusterId(),
		DestinationClusterId: link.GetDestinationClusterId(),
		RemoteClusterId:      link.GetRemoteClusterId(),
		State:                link.GetLinkState(),
		Error:                linkError,
		ErrorMessage:         link.GetLinkErrorMessage(),
	}
}

func (c *linkCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List cluster links.",
		Long:  "List cluster links if the provided cluster is a destination cluster.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	cmd.Flags().Bool(includeTopicsFlagName, false, "If set, will list mirrored topics for the links returned.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *linkCommand) list(cmd *cobra.Command, _ []string) error {
	includeTopics, err := cmd.Flags().GetBool(includeTopicsFlagName)
	if err != nil {
		return err
	}

	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return err
	}

	links, err := kafkaREST.CloudClient.ListKafkaLinks()
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, link := range links {
		if includeTopics {
			for _, topic := range link.GetTopicNames() {
				list.Add(newLink(link, topic))
			}
		} else {
			list.Add(newLink(link, ""))
		}
	}
	list.Filter(getListFields(includeTopics))
	return list.Print()
}

func getListFields(includeTopics bool) []string {
	x := []string{"Name"}

	if includeTopics {
		x = append(x, "TopicName")
	}

	return append(x, "SourceClusterId", "DestinationClusterId", "RemoteClusterId", "State", "Error", "ErrorMessage")
}
