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

type task struct {
	TaskName string
	State    string
	Errors   []taskErr
}

type taskErr struct {
	ErrorCode    string
	ErrorMessage string
}

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

func (c *linkCommand) newTaskCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Manager a cluster link's tasks.",
	}

	cmd.AddCommand(c.newLinkTaskListCommand())

	return cmd
}

func (c *linkCommand) newLinkTaskListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "list <link>",
		Short:             "List a cluster link's tasks.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.taskList,
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *linkCommand) taskList(cmd *cobra.Command, args []string) error {
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
	tasks := make([]task, len(link.GetTasks()))
	for i, t := range link.GetTasks() {
		errs := make([]taskErr, len(t.Errors))
		for j, e := range t.Errors {
			errs[j] = taskErr{
				ErrorCode:    e.ErrorCode,
				ErrorMessage: e.ErrorMessage,
			}
		}
		tasks[i] = task{
			TaskName: t.TaskName,
			State: t.State,
			Errors:   errs,
		}
	}

	isSerialized := output.GetFormat(cmd).IsSerialized()
	if isSerialized {
		return writeSerialized(cmd, tasks)
	} else {
		return writeHuman(cmd, tasks)
	}
}

func writeSerialized(cmd *cobra.Command, tasks []task) error {
	list := output.NewList(cmd)
	for _, task := range tasks {
		errs := make([]serializedTaskErrorOut, len(task.Errors))
		for i, err := range task.Errors {
			errs[i] = serializedTaskErrorOut(err)
		}
		list.Add(&serializedTaskOut{
			TaskName: task.TaskName,
			State:    task.State,
			Errors:   errs,
		})
	}
	return list.Print()
}

func writeHuman(cmd *cobra.Command, tasks []task) error {
	list := output.NewList(cmd)
	for _, task := range tasks {
		errs := make([]string, len(task.Errors))
		for i, err := range task.Errors {
			errs[i] = fmt.Sprintf(`%s: "%s"`, err.ErrorCode, err.ErrorMessage)
		}
		// Encode the list of errors into a single, comma separated String so that the errors take up a single column
		// in the outputted table.
		errsStr := strings.Join(errs, ", ")
		list.Add(&humanTaskOut{
			TaskName: task.TaskName,
			State:    task.State,
			Errors:   errsStr,
		})
	}
	return list.Print()
}
