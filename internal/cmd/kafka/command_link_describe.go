package kafka

import (
	"github.com/spf13/cobra"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type describeOut struct {
	Name                 string `human:"Name" serialized:"link_name"`
	TopicName            string `human:"Topic Name" serialized:"topic_name"`
	SourceClusterId      string `human:"Source Cluster" serialized:"source_cluster_id"`
	DestinationClusterId string `human:"Destination Cluster" serialized:"destination_cluster_id"`
	RemoteClusterId      string `human:"Remote Cluster" serialized:"remote_cluster_id"`
	State                string `human:"State" serialized:"state"`
	Error                string `human:"Error,omitempty" serialized:"error,omitempty"`
	ErrorMessage         string `human:"Error Message,omitempty" serialized:"error_message,omitempty"`
}

func (c *linkCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <link>",
		Short:             "Describe a cluster link.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describe,
	}

	pcmd.AddForceFlag(cmd)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *linkCommand) describe(cmd *cobra.Command, args []string) error {
	linkName := args[0]

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

	data, httpResp, err := kafkaREST.CloudClient.GetKafkaLink(clusterId, linkName)
	if err != nil {
		return kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
	}

	table := output.NewTable(cmd)
	table.Add(newDescribeLink(data, ""))
	table.Filter(getListFields(false))
	return table.Print()
}

func newDescribeLink(data kafkarestv3.ListLinksResponseData, topic string) *describeOut {
	var linkError string
	if data.GetLinkError() != "NO_ERROR" {
		linkError = data.GetLinkError()
	}
	return &describeOut{
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
