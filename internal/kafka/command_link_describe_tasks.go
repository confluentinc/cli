package kafka

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/kafkarest"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type serializedTaskOut struct {
	TaskName string                   `serialized:"task_name"`
	State    string                   `serialized:"state"`
	Errors   []serializedTaskErrorOut `serialized:"errors"`
}

type humanTaskOut struct {
	TaskName string `human:"Task Name"`
	State    string `human:"State"`
	Errors   string `human:"Errors"`
}

type serializedTaskErrorOut struct {
	ErrorCode    string `serialized:"error_code"`
	ErrorMessage string `serialized:"error_message"`
}

func (c *linkCommand) newDescribeLinkTasksCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe-tasks <link>",
		Short:             "Describe a cluster links tasks.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describeTasks,
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *linkCommand) describeTasks(cmd *cobra.Command, args []string) error {
	linkName := args[0]

	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return err
	}

	apiContext := context.WithValue(context.Background(), kafkarestv3.ContextAccessToken, kafkaREST.CloudClient.AuthToken)
	req := kafkaREST.CloudClient.ClusterLinkingV3Api.GetKafkaLink(apiContext, kafkaREST.CloudClient.ClusterId, linkName)
	req = req.IncludeTasks(true)
	res, httpResp, err := req.Execute()
	link, err := res, kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
	if err != nil {
		return err
	}

	describeTasksOut := newDescribeTasksLink(link)
	isSerialized := output.GetFormat(cmd).IsSerialized()
	if isSerialized {
		list := output.NewList(cmd)
		for _, taskOut := range describeTasksOut {
			list.Add(&taskOut)
		}
		return list.Print()
	} else {
		return printHumanTaskOuts(cmd, describeTasksOut)
	}
}

func printHumanTaskOuts(cmd *cobra.Command, taskOuts []serializedTaskOut) error {
	list := output.NewList(cmd)
	for _, taskOut := range taskOuts {
		errs := make([]string, 0)
		for _, err := range taskOut.Errors {
			errs = append(errs, fmt.Sprintf("Error Code: %s Error Message: %s", err.ErrorCode, err.ErrorMessage))
		}
		// Encode the list of errors into a single, comma separated String so that the errors take up a single column
		// in the outputted table.
		errsStr := strings.Join(errs, ",")
		list.Add(&humanTaskOut{
			TaskName: taskOut.TaskName,
			State:    taskOut.State,
			Errors:   errsStr,
		})
	}
	return list.Print()
}

func newDescribeTasksLink(link kafkarestv3.ListLinksResponseData) []serializedTaskOut {
	tasks := link.GetTasks()
	if tasks == nil {
		return make([]serializedTaskOut, 0)
	}
	taskOuts := make([]serializedTaskOut, 0)
	for _, task := range tasks {
		taskErrorOuts := make([]serializedTaskErrorOut, 0)
		for _, err := range task.Errors {
			taskErrorOuts = append(taskErrorOuts, serializedTaskErrorOut{
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