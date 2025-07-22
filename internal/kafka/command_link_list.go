package kafka

import (
	"github.com/spf13/cobra"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

const includeTopicsFlagName = "include-topics"

func newLink(link kafkarestv3.ListLinksResponseData, topic string) *linkOut {
	var linkError string
	if link.GetLinkError() != "NO_ERROR" {
		linkError = link.GetLinkError()
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
	pcmd.AddEndpointFlag(cmd, c.AuthenticatedCLICommand)
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

	kafkaREST, err := c.GetKafkaREST(cmd)
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

	return append(x, "SourceCluster", "DestinationCluster", "RemoteCluster", "State", "Error", "ErrorMessage")
}
