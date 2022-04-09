package kafka

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"
)

func (c *linkCommand) listOnPrem(cmd *cobra.Command, args []string) error {
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
		return kafkaRestError(pcmd.GetCPKafkaRestBaseUrl(client), err, httpResp)
	}

	listFields := getListFields(includeTopics, c.cfg.IsCloudLogin())
	humanLabels := camelToSpaced(listFields)
	structuredLabels := camelToSnake(listFields)

	w, err := output.NewListOutputWriter(cmd, listFields, humanLabels, structuredLabels)
	if err != nil {
		return err
	}

	for _, data := range listLinksRespDataList.Data {
		if includeTopics && len(data.TopicsNames) > 0 {
			for _, topic := range data.TopicsNames {
				w.AddElement(newCPLink(data, topic))
			}
		} else {
			w.AddElement(newCPLink(data, ""))
		}
	}

	return w.Out()
}

func newCPLink(data kafkarestv3.ListLinksResponseData, topic string) *link {
	l := &link{
		LinkName:  data.LinkName,
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
