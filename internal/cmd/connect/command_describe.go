package connect

import (
	"github.com/spf13/cobra"

	connectv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect/v1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type taskDescribeOut struct {
	TaskId int32  `human:"Task ID" serialized:"task_id"`
	State  string `human:"State" serialized:"state"`
}

type configDescribeOut struct {
	Config string `human:"Config" serialized:"config"`
	Value  string `human:"Value" serialized:"value"`
}

type structuredDescribeDisplay struct {
	Connector *connectOut         `serialized:"connector"`
	Tasks     []taskDescribeOut   `serialized:"task"`
	Configs   []configDescribeOut `serialized:"configs"`
}

func (c *command) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe a connector.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describe,
		Annotations:       map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe connector and task level details of a connector in the current or specified Kafka cluster context.",
				Code: "confluent connect describe lcc-123456",
			},
			examples.Example{
				Code: "confluent connect describe lcc-123456 --cluster lkc-123456",
			},
		),
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) describe(cmd *cobra.Command, args []string) error {
	kafkaCluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	connectorExpansion, err := c.V2Client.GetConnectorExpansionById(args[0], c.EnvironmentId(), kafkaCluster.ID)
	if err != nil {
		return err
	}

	outputOption, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	}

	if outputOption == output.Human.String() {
		return printHumanDescribe(cmd, connectorExpansion)
	}

	return printStructuredDescribe(cmd, connectorExpansion)
}

func printHumanDescribe(cmd *cobra.Command, connector *connectv1.ConnectV1ConnectorExpansion) error {
	utils.Println(cmd, "Connector Details")
	table := output.NewTable(cmd)
	table.Add(&connectOut{
		Name:   connector.Status.GetName(),
		Id:     connector.Id.GetId(),
		Status: connector.Status.Connector.GetState(),
		Type:   connector.Status.GetType(),
		Trace:  connector.Status.Connector.GetTrace(),
	})
	if err := table.Print(); err != nil {
		return err
	}

	utils.Println(cmd, "\n\nTask Level Details")
	list := output.NewList(cmd)
	for _, task := range connector.Status.GetTasks() {
		list.Add(&taskDescribeOut{task.Id, task.State})
	}
	if err := list.Print(); err != nil {
		return err
	}

	utils.Println(cmd, "\n\nConfiguration Details")
	list = output.NewList(cmd)
	for name, value := range connector.Info.GetConfig() {
		list.Add(&configDescribeOut{name, value})
	}
	return list.Print()
}

func printStructuredDescribe(cmd *cobra.Command, connector *connectv1.ConnectV1ConnectorExpansion) error {
	tasks := make([]taskDescribeOut, 0)
	for _, task := range connector.Status.GetTasks() {
		tasks = append(tasks, taskDescribeOut{TaskId: task.Id, State: task.State})
	}

	configs := make([]configDescribeOut, 0)
	for name, value := range connector.Info.GetConfig() {
		configs = append(configs, configDescribeOut{Config: name, Value: value})
	}

	table := output.NewTable(cmd)
	table.Add(&structuredDescribeDisplay{
		Connector: &connectOut{
			Name:   connector.Status.GetName(),
			Id:     connector.Id.GetId(),
			Status: connector.Status.Connector.GetState(),
			Type:   connector.Status.GetType(),
			Trace:  connector.Status.Connector.GetTrace(),
		},
		Tasks:   tasks,
		Configs: configs,
	})
	return table.Print()
}
