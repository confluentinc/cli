package kafka

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func newLinkOnPrem(data kafkarestv3.ListLinksResponseData, topic string) *link {
	l := &link{
		Name:      data.LinkName,
		TopicName: topic,
	}

	if data.SourceClusterId != nil {
		l.SourceClusterId = *data.SourceClusterId
	}

	if data.DestinationClusterId != nil {
		l.DestinationClusterId = *data.DestinationClusterId
	}

	return l
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

	client, ctx, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	clusterId, err := getClusterIdForRestRequests(client, ctx)
	if err != nil {
		return err
	}

	listLinksRespDataList, httpResp, err := client.ClusterLinkingV3Api.ListKafkaLinks(ctx, clusterId)
	if err != nil {
		return handleOpenApiError(httpResp, err, client)
	}

	list := output.NewList(cmd)
	for _, data := range listLinksRespDataList.Data {
		if includeTopics {
			if len(data.TopicNames) > 0 {
				for _, topic := range data.TopicNames {
					list.Add(newLinkOnPrem(data, topic))
				}
			}
		} else {
			list.Add(newLinkOnPrem(data, ""))
		}
	}
	list.Filter(getListFieldsOnPrem(includeTopics))
	return list.Print()
}
