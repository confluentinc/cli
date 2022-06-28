package kafka

import (
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
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

func (c *linkCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List previously created cluster links.",
		Long: "List previously created cluster links if the provided cluster is a destination cluster.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	cmd.Flags().Bool(includeTopicsFlagName, false, "If set, will list mirrored topics for the links returned.")

	if c.cfg.IsCloudLogin() {
		pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *linkCommand) list(cmd *cobra.Command, _ []string) error {
	includeTopics, err := cmd.Flags().GetBool(includeTopicsFlagName)
	if err != nil {
		return err
	}

	client, ctx, clusterId, err := c.getKafkaRestComponents(cmd)
	if err != nil {
		return err
	}

	listLinksRespDataList, httpResp, err := client.ClusterLinkingV3Api.ListKafkaLinks(ctx, clusterId)
	if err != nil {
		return handleOpenApiError(httpResp, err, client)
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
				w.AddElement(newLink(data, topic))
			}
		} else {
			w.AddElement(newLink(data, ""))
		}
	}

	return w.Out()
}

func getListFields(includeTopics, isCloud bool) []string {
	x := []string{"LinkName"}

	if includeTopics {
		x = append(x, "TopicName")
	}

	if isCloud {
		x = append(x, "SourceClusterId")
	} else {
		x = append(x, "DestinationClusterId")
	}

	return x
}
