package kafka

import (
	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

const includeTopicsFlagName = "include-topics"

type link struct {
	LinkName             string
	TopicName            string
	SourceClusterId      string
	DestinationClusterId string
}

func newLink(data kafkarestv3.ListLinksResponseData, topic string) *link {
	return &link{
		LinkName:             data.LinkName,
		TopicName:            topic,
		SourceClusterId:      data.GetSourceClusterId(),
		DestinationClusterId: data.GetDestinationClusterId(),
	}
}

func (c *linkCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List previously created cluster links.",
		Long:  "List previously created cluster links if the provided cluster is a destination cluster.",
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
	if kafkaREST == nil {
		if err != nil {
			return err
		}
		return errors.New(errors.RestProxyNotAvailableMsg)
	}

	clusterId, err := getKafkaClusterLkcId(c.AuthenticatedStateFlagCommand)
	if err != nil {
		return err
	}

	listLinksRespDataList, httpResp, err := kafkaREST.CloudClient.ListKafkaLinks(clusterId)
	if err != nil {
		return kafkaRestError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
	}

	listFields := getListFields(includeTopics)
	humanLabels := camelToSpaced(listFields)
	structuredLabels := camelToSnake(listFields)

	w, err := output.NewListOutputWriter(cmd, listFields, humanLabels, structuredLabels)
	if err != nil {
		return err
	}

	for _, data := range listLinksRespDataList.Data {
		if includeTopics && len(*data.TopicsNames) > 0 {
			for _, topic := range *data.TopicsNames {
				w.AddElement(newLink(data, topic))
			}
		} else {
			w.AddElement(newLink(data, ""))
		}
	}

	return w.Out()
}

func getListFields(includeTopics bool) []string {
	x := []string{"LinkName"}

	if includeTopics {
		x = append(x, "TopicName")
	}

	return append(x, "SourceClusterId", "DestinationClusterId")
}
