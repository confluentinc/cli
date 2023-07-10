package kafka

import (
	"github.com/spf13/cobra"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/output"
)

const includeTopicsFlagName = "include-topics"

type link struct {
	Name                 string `human:"Name" serialized:"link_name"`
	TopicName            string `human:"Topic Name" serialized:"topic_name"`
	SourceClusterId      string `human:"Source Cluster" serialized:"source_cluster_id"`
	DestinationClusterId string `human:"Destination Cluster" serialized:"destination_cluster_id"`
	RemoteClusterId      string `human:"Remote Cluster" serialized:"remote_cluster_id"`
	State                string `human:"State" serialized:"state"`
	Error                string `human:"Error" serialized:"error"`
	ErrorMessage         string `human:"Error Message" serialized:"error_message"`
}

func newLink(data kafkarestv3.ListLinksResponseData, topic string) *link {
	var linkError string
	if data.GetLinkError() != "NO_ERROR" {
		linkError = data.GetLinkError()
	}
	return &link{
		Name:                 data.LinkName,
		TopicName:            topic,
		SourceClusterId:      data.GetSourceClusterId(),
		DestinationClusterId: data.GetDestinationClusterId(),
		RemoteClusterId:      data.GetRemoteClusterId(),
		State:                data.GetLinkState(),
		Error:                linkError,
		ErrorMessage:         data.GetLinkErrorMessage(),
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
	if kafkaREST == nil {
		if err != nil {
			return err
		}
		return errors.New(errors.RestProxyNotAvailableMsg)
	}

	clusterId, err := getKafkaClusterLkcId(c.AuthenticatedCLICommand)
	if err != nil {
		return err
	}

	listLinksRespDataList, httpResp, err := kafkaREST.CloudClient.ListKafkaLinks(clusterId)
	if err != nil {
		return kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
	}

	list := output.NewList(cmd)
	for _, data := range listLinksRespDataList.Data {
		if includeTopics {
			// data.GetTopicsNames() is empty even when the http response contains a non-empty list of topic names,
			// this function call is a temporary work-around for this issue
			mirrorTopicNames, err := getMirrorTopicNames(kafkaREST, clusterId, data.GetLinkName())
			if err != nil {
				return err
			}

			for _, topicName := range mirrorTopicNames {
				list.Add(newLink(data, topicName))
			}
		} else {
			list.Add(newLink(data, ""))
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

func getMirrorTopicNames(kafkaREST *pcmd.KafkaREST, clusterId, linkName string) ([]string, error) {
	listMirrorTopicsResponseDataList, httpResp, err := kafkaREST.CloudClient.ListKafkaMirrorTopicsUnderLink(clusterId, linkName)
	if err != nil {
		return nil, kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
	}

	mirrorTopicNames := make([]string, len(listMirrorTopicsResponseDataList.GetData()))
	for i, data := range listMirrorTopicsResponseDataList.GetData() {
		mirrorTopicNames[i] = data.GetMirrorTopicName()
	}

	return mirrorTopicNames, nil
}
