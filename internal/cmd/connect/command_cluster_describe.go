package connect

import (
	"github.com/spf13/cobra"

	connectv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect/v1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type serializedDescribeOut struct {
	Connector *serializedConnectorOut `json:"connector" yaml:"connector"`
	Tasks     []serializedTasksOut    `json:"tasks" yaml:"tasks"`
	Configs   []serializedConfigsOut  `json:"configs" yaml:"configs"`
}

type serializedConnectorOut struct {
	Id     string `json:"id" yaml:"id"`
	Name   string `json:"name" yaml:"name"`
	Status string `json:"status" yaml:"status"`
	Type   string `json:"type" yaml:"type"`
	Trace  string `json:"trace,omitempty" yaml:"trace,omitempty"`
}

type taskDescribeOut struct {
	TaskId int32  `human:"Task ID"`
	State  string `human:"State"`
}

type serializedTasksOut struct {
	TaskId int32  `json:"task_id" yaml:"task_id"`
	State  string `json:"state" yaml:"state"`
}

type configDescribeOut struct {
	Config string `human:"Config"`
	Value  string `human:"Value"`
}

type serializedConfigsOut struct {
	Config string `json:"config" yaml:"config"`
	Value  string `json:"value" yaml:"value"`
}

func (c *clusterCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id|name>",
		Short:             "Describe a connector.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describe,
		Annotations:       map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe connector and task level details of a connector in the current or specified Kafka cluster context.",
				Code: "confluent connect cluster describe lcc-123456",
			},
			examples.Example{
				Code: "confluent connect cluster describe lcc-123456 --cluster lkc-123456",
			},
		),
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *clusterCommand) describe(cmd *cobra.Command, args []string) error {
	kafkaCluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	connector, err := c.V2Client.GetConnectorExpansionById(args[0], environmentId, kafkaCluster.ID)
	if err != nil {
		if connector, err = c.V2Client.GetConnectorExpansionByName(args[0], environmentId, kafkaCluster.ID); err != nil {
			return err
		}
	}

	if output.GetFormat(cmd) == output.Human {
		return printHumanDescribe(cmd, connector)
	}

	return printSerializedDescribe(cmd, connector)
}

func printHumanDescribe(cmd *cobra.Command, connector *connectv1.ConnectV1ConnectorExpansion) error {
	output.Println("Connector Details")
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
	output.Println()
	output.Println()

	output.Println("Task Level Details")
	list := output.NewList(cmd)
	for _, task := range connector.Status.GetTasks() {
		list.Add(&taskDescribeOut{
			TaskId: task.GetId(),
			State:  task.GetState(),
		})
	}
	if err := list.Print(); err != nil {
		return err
	}
	output.Println()
	output.Println()

	output.Println("Configuration Details")
	list = output.NewList(cmd)
	for name, value := range connector.Info.GetConfig() {
		list.Add(&configDescribeOut{
			Config: name,
			Value:  value,
		})
	}
	return list.Print()
}

func printSerializedDescribe(cmd *cobra.Command, connector *connectv1.ConnectV1ConnectorExpansion) error {
	tasks := make([]serializedTasksOut, 0)
	for _, task := range connector.Status.GetTasks() {
		tasks = append(tasks, serializedTasksOut{TaskId: task.Id, State: task.State})
	}

	configs := make([]serializedConfigsOut, 0)
	for name, value := range connector.Info.GetConfig() {
		configs = append(configs, serializedConfigsOut{Config: name, Value: value})
	}

	out := &serializedDescribeOut{
		Connector: &serializedConnectorOut{
			Id:     connector.Id.GetId(),
			Name:   connector.Status.GetName(),
			Status: connector.Status.Connector.GetState(),
			Type:   connector.Status.GetType(),
			Trace:  connector.Status.Connector.GetTrace(),
		},
		Tasks:   tasks,
		Configs: configs,
	}
	return output.SerializedOutput(cmd, out)
}
