package kafka

import (
	"bytes"
	"context"

	"github.com/spf13/cobra"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/kafkarest"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type describeOut struct {
	Name                 string              `human:"Name" serialized:"link_name"`
	TopicName            string              `human:"Topic Name" serialized:"topic_name"`
	SourceClusterId      string              `human:"Source Cluster" serialized:"source_cluster_id"`
	DestinationClusterId string              `human:"Destination Cluster" serialized:"destination_cluster_id"`
	RemoteClusterId      string              `human:"Remote Cluster" serialized:"remote_cluster_id"`
	State                string              `human:"State" serialized:"state"`
	Error                string              `human:"Error,omitempty" serialized:"error,omitempty"`
	ErrorMessage         string              `human:"Error Message,omitempty" serialized:"error_message,omitempty"`
	Tasks                []serializedTaskOut `serialized:"tasks"`
}

type serializedTaskOut struct {
	TaskName string         `serialized:"task_name"`
	State    string         `serialized:"state"`
	Errors   []taskErrorOut `serialized:"errors"`
}

type humanTaskOut struct {
	TaskName string `human:"Task name"`
	State    string `human:"state"`
	Errors   string `human:"errors"`
}

type taskErrorOut struct {
	ErrorCode    string `human:"Error code" serialized:"error_code"`
	ErrorMessage string `human:"Error message" serialized:"error_message"`
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
	describeOut := newDescribeLink(link, "")
	table.Add(describeOut)
	isSerialized := output.GetFormat(cmd).IsSerialized()
	if isSerialized {
		// If we are serializing the output then there's no need to do any customization of the output. It will get
		// correctly serialized.
		table.Filter(getDescribeClusterLinksFields(true))
		return table.Print()
	} else {
		// If we are not serializing the output, which means it's for human consumption, then we do some customization
		// so it's more readable.
		// The main link info gets output in table format which means it has two columns. Because there are multiple
		// tasks, and each task itself can have multiple errors it's awkward to shove all the output into a single
		// column. As a result we print a separate list dedicated to the tasks after the first table that contains the
		// main link information.
		table.Filter(getDescribeClusterLinksFields(false))
		if err != nil {
			return err
		}
		err = table.Print()
		if err != nil {
			return err
		}
		taskOuts := describeOut.Tasks
		if len(taskOuts) > 1 {
			list := output.NewList(cmd)
			for i := range taskOuts {
				t := taskOuts[i]
				var errsStr bytes.Buffer
				for i := range t.Errors {
					eo := t.Errors[i]
					errsStr.WriteString("Error code: ")
					errsStr.WriteString(eo.ErrorCode)
					errsStr.WriteString(" ")
					errsStr.WriteString("Error message: ")
					errsStr.WriteString(eo.ErrorMessage)
					if i < len(t.Errors)-1 {
						errsStr.WriteString(",")
					}
				}
				list.Add(&humanTaskOut{
					TaskName: t.TaskName,
					State:    t.State,
					Errors:   errsStr.String(),
				})
			}
			return list.Print()
		} else {
			return nil
		}
	}
}

func newDescribeLink(link kafkarestv3.ListLinksResponseData, topic string) *describeOut {
	var linkError string
	if link.GetLinkError() != "NO_ERROR" {
		linkError = link.GetLinkError()
	}
	tasks := toTaskOut(link.GetTasks())
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
	}
}

func toTaskOut(tasks []kafkarestv3.LinkTask) []serializedTaskOut {
	var tasksToEncode []kafkarestv3.LinkTask
	if tasks != nil {
		tasksToEncode = tasks
	} else {
		tasksToEncode = make([]kafkarestv3.LinkTask, 0)
	}
	taskOuts := make([]serializedTaskOut, 0)
	for _, task := range tasksToEncode {
		taskErrorOuts := make([]taskErrorOut, 0)
		for _, err := range task.Errors {
			taskErrorOuts = append(taskErrorOuts, taskErrorOut{
				ErrorCode:    err.ErrorCode,
				ErrorMessage: err.ErrorMessage,
			})
		}
		taskOuts = append(taskOuts, serializedTaskOut{
			TaskName: task.TaskName,
			State:    task.State,
			Errors:   taskErrorOuts,
		})
	}
	return taskOuts
}

func getDescribeClusterLinksFields(includeTasks bool) []string {
	x := []string{"Name", "SourceClusterId", "DestinationClusterId", "RemoteClusterId", "State", "Error", "ErrorMessage"}
	if includeTasks {
		x = append(x, "Tasks", "TaskName", "Errors", "ErrorCode")
	}
	return x
}
