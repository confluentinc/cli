package kafka

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/spf13/cobra"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/kafkarest"
	"github.com/confluentinc/cli/v3/pkg/output"
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
	Tasks                string `human:"Tasks,omitempty" serialized:"tasks,omitempty"`
}

func (c *linkCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <link>",
		Short:             "Describe a cluster link.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describe,
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *linkCommand) describe(cmd *cobra.Command, args []string) error {
	linkName := args[0]

	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return err
	}

	cloudClient := kafkaREST.CloudClient
	apiContext := context.WithValue(context.Background(), kafkarestv3.ContextAccessToken, cloudClient.AuthToken)
	req := cloudClient.ClusterLinkingV3Api.GetKafkaLink(apiContext, cloudClient.ClusterId, linkName)
	req = req.IncludeTasks(true)
	res, httpResp, err := req.Execute()
	link, err := res, kafkarest.NewError(cloudClient.GetUrl(), err, httpResp)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	describeOut, err := newDescribeLink(link, "")
	if err != nil {
		return err
	}
	table.Add(describeOut)
	table.Filter(getDescribeClusterLinksFields())
	return table.Print()
}

func newDescribeLink(link kafkarestv3.ListLinksResponseData, topic string) (*describeOut, error) {
	var linkError string
	if link.GetLinkError() != "NO_ERROR" {
		linkError = link.GetLinkError()
	}
	tasks, err := toTaskOut(link.GetTasks())
	if err != nil {
		return nil, err
	}
	return &describeOut{
		Name:                 link.GetLinkName(),
		TopicName:            topic,
		SourceClusterId:      link.GetSourceClusterId(),
		DestinationClusterId: link.GetDestinationClusterId(),
		RemoteClusterId:      link.GetRemoteClusterId(),
		State:                link.GetLinkState(),
		Error:                linkError,
		ErrorMessage:         link.GetLinkErrorMessage(),
		Tasks:                tasks,
	}, nil
}

func toTaskOut(tasks []kafkarestv3.LinkTask) (string, error) {
	var tasksToEncode []kafkarestv3.LinkTask
	if tasks != nil {
		tasksToEncode = tasks
	} else {
		// If nil create an empty slice so that the encoded JSON is [] instead of null.
		tasksToEncode = make([]kafkarestv3.LinkTask, 0)
	}
	var b bytes.Buffer
	err := json.NewEncoder(&b).Encode(tasksToEncode)
	if err != nil {
		return "", err
	} else {
		return b.String(), nil
	}
}

func getDescribeClusterLinksFields() []string {
	x := []string{"Name", "SourceClusterId", "DestinationClusterId", "RemoteClusterId", "State", "Error", "ErrorMessage", "Tasks"}
	return x
}
