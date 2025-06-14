package kafka

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func newLinkOnPrem(data kafkarestv3.ListLinksResponseData, topic string) *linkOut {
	listEntry := &linkOut{
		Name:      data.LinkName,
		TopicName: topic,
		State:     data.LinkState,
	}

	if data.SourceClusterId != nil {
		listEntry.SourceCluster = *data.SourceClusterId
	}
	if data.DestinationClusterId != nil {
		listEntry.DestinationCluster = *data.DestinationClusterId
	}
	if data.RemoteClusterId != nil {
		listEntry.RemoteCluster = *data.RemoteClusterId
	}
	if data.LinkError != "NO_ERROR" {
		listEntry.Error = data.LinkError
	}
	if data.LinkErrorMessage != nil {
		listEntry.ErrorMessage = *data.LinkErrorMessage
	}

	return listEntry
}

func (c *linkCommand) newListCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List cluster links.",
		Long:  "List cluster links if the provided cluster is a destination cluster.",
		Args:  cobra.NoArgs,
		RunE:  c.listOnPrem,
	}

	cmd.Flags().Bool(includeTopicsFlagName, false, "If set, will list mirrored topics for the links returned.")
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *linkCommand) listOnPrem(cmd *cobra.Command, _ []string) error {
	includeTopics, err := cmd.Flags().GetBool(includeTopicsFlagName)
	if err != nil {
		return err
	}

	client, ctx, clusterId, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	links, httpResp, err := client.ClusterLinkingV3Api.ListKafkaLinks(ctx, clusterId)
	if err != nil {
		return handleOpenApiError(httpResp, err, client)
	}

	list := output.NewList(cmd)
	for _, link := range links.Data {
		if includeTopics {
			for _, topic := range link.TopicNames {
				list.Add(newLinkOnPrem(link, topic))
			}
		} else {
			list.Add(newLinkOnPrem(link, ""))
		}
	}
	list.Filter(getListFieldsOnPrem(includeTopics))
	return list.Print()
}
