package connect

import (
	"os"

	"github.com/confluentinc/go-printer"
	"github.com/spf13/cobra"

	connectv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect/v1"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type taskDescribeDisplay struct {
	TaskId int32  `json:"task_id" yaml:"task_id"`
	State  string `json:"state" yaml:"state"`
}

type configDescribeDisplay struct {
	Config string `json:"config" yaml:"config"`
	Value  string `json:"value" yaml:"value"`
}

type structuredDescribeDisplay struct {
	Connector *connectorDescribeDisplay `json:"connector" yaml:"connector"`
	Tasks     []taskDescribeDisplay     `json:"tasks" yaml:"task"`
	Configs   []configDescribeDisplay   `json:"configs" yaml:"configs"`
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
		printHumanDescribe(cmd, connectorExpansion)
		return nil
	}

	return printStructuredDescribe(connectorExpansion, outputOption)
}

func printHumanDescribe(cmd *cobra.Command, connector *connectv1.ConnectV1ConnectorExpansion) {
	utils.Println(cmd, "Connector Details")
	data := &connectorDescribeDisplay{
		Name:   connector.Status.GetName(),
		ID:     connector.Id.GetId(),
		Status: connector.Status.Connector.GetState(),
		Type:   connector.Status.GetType(),
		Trace:  connector.Status.Connector.GetTrace(),
	}
	_ = printer.RenderTableOut(data, listFields, map[string]string{}, os.Stdout)

	utils.Println(cmd, "\n\nTask Level Details")
	var tasks [][]string
	for _, task := range connector.Status.GetTasks() {
		row := printer.ToRow(&taskDescribeDisplay{task.Id, task.State}, []string{"TaskId", "State"})
		tasks = append(tasks, row)
	}
	printer.RenderCollectionTable(tasks, []string{"Task ID", "State"})

	utils.Println(cmd, "\n\nConfiguration Details")
	var configs [][]string
	titleRow := []string{"Config", "Value"}
	for name, value := range connector.Info.GetConfig() {
		row := printer.ToRow(&configDescribeDisplay{name, value}, titleRow)
		configs = append(configs, row)
	}
	printer.RenderCollectionTable(configs, titleRow)
}

func printStructuredDescribe(connector *connectv1.ConnectV1ConnectorExpansion, format string) error {
	structuredDisplay := &structuredDescribeDisplay{
		Connector: &connectorDescribeDisplay{
			Name:   connector.Status.GetName(),
			ID:     connector.Id.GetId(),
			Status: connector.Status.Connector.GetState(),
			Type:   connector.Status.GetType(),
			Trace:  connector.Status.Connector.GetTrace(),
		},
		Tasks:   []taskDescribeDisplay{},
		Configs: []configDescribeDisplay{},
	}
	for _, task := range connector.Status.GetTasks() {
		structuredDisplay.Tasks = append(structuredDisplay.Tasks, taskDescribeDisplay{
			TaskId: task.Id,
			State:  task.State,
		})
	}
	for name, value := range connector.Info.GetConfig() {
		structuredDisplay.Configs = append(structuredDisplay.Configs, configDescribeDisplay{
			Config: name,
			Value:  value,
		})
	}
	return output.StructuredOutput(format, structuredDisplay)
}
