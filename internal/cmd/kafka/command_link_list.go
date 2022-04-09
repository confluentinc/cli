package kafka

import (
	cloudkafkarest "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
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

func newCloudLink(data cloudkafkarest.ListLinksResponseData, topic string) *link {
	l := &link{
		LinkName:  data.LinkName,
		TopicName: topic,
	}

	if data.SourceClusterId != nil {
		l.SourceClusterId = *data.SourceClusterId
	}

	return l
}

func (c *linkCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List previously created cluster links.",
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

	kafkaREST, err := c.GetCloudKafkaREST()
	if err != nil {
		return err
	}

	kafkaClusterConfig, err := c.AuthenticatedCLICommand.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}
	clusterId := kafkaClusterConfig.ID

	listLinksRespDataList, httpResp, err := kafkaREST.Client.ClusterLinkingV3Api.ListKafkaLinks(kafkaREST.Context, clusterId).Execute()
	if err != nil {
		return kafkaRestError(pcmd.GetCloudKafkaRestBaseUrl(kafkaREST.Client), err, httpResp)
	}

	listFields := getListFields(includeTopics, c.cfg.IsCloudLogin())
	humanLabels := camelToSpaced(listFields)
	structuredLabels := camelToSnake(listFields)

	w, err := output.NewListOutputWriter(cmd, listFields, humanLabels, structuredLabels)
	if err != nil {
		return err
	}

	for _, data := range listLinksRespDataList.Data {
		if includeTopics && len(*data.TopicsNames) > 0 {
			for _, topic := range *data.TopicsNames {
				w.AddElement(newCloudLink(data, topic))
			}
		} else {
			w.AddElement(newCloudLink(data, ""))
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
