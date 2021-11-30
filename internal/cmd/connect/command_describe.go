package connect

import (
	"context"
	"os"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	opv1 "github.com/confluentinc/cc-structs/operator/v1"
	"github.com/confluentinc/go-printer"
	"github.com/spf13/cobra"

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
		Use:         "describe <id>",
		Short:       "Describe a connector.",
		Args:        cobra.ExactArgs(1),
		RunE:        pcmd.NewCLIRunE(c.describe),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
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

	cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)

	return cmd
}

func (c *command) describe(cmd *cobra.Command, args []string) error {
	kafkaCluster, err := c.Context.GetKafkaClusterForCommand(cmd)
	if err != nil {
		return err
	}

	connector := &schedv1.Connector{
		AccountId:      c.EnvironmentId(),
		KafkaClusterId: kafkaCluster.ID,
		Id:             args[0],
	}

	connectorExpansion, err := c.Client.Connect.GetExpansionById(context.Background(), connector)
	if err != nil {
		return err
	}

	outputOption, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	}

	if outputOption == output.Human.String() {
		return printHumanDescribe(cmd, connectorExpansion)
	} else {
		return printStructuredDescribe(connectorExpansion, outputOption)
	}
}

func printHumanDescribe(cmd *cobra.Command, connector *opv1.ConnectorExpansion) error {
	utils.Println(cmd, "Connector Details")
	data := &connectorDescribeDisplay{
		Name:   connector.Status.Name,
		ID:     connector.Id.Id,
		Status: connector.Status.Connector.State,
		Type:   connector.Info.Type,
		Trace:  connector.Status.Connector.Trace,
	}
	_ = printer.RenderTableOut(data, listFields, map[string]string{}, os.Stdout)

	utils.Println(cmd, "\n\nTask Level Details")
	var tasks [][]string
	titleRow := []string{"TaskId", "State"}
	for _, task := range connector.Status.Tasks {
		tasks = append(tasks, printer.ToRow(&taskDescribeDisplay{
			task.Id,
			task.State,
		}, titleRow))
	}
	printer.RenderCollectionTable(tasks, titleRow)
	utils.Println(cmd, "\n\nConfiguration Details")
	var configs [][]string
	titleRow = []string{"Config", "Value"}
	for name, value := range connector.Info.Config {
		configs = append(configs, printer.ToRow(&configDescribeDisplay{
			name,
			value,
		}, titleRow))
	}
	printer.RenderCollectionTable(configs, titleRow)
	return nil
}

func printStructuredDescribe(connector *opv1.ConnectorExpansion, format string) error {
	structuredDisplay := &structuredDescribeDisplay{
		Connector: &connectorDescribeDisplay{
			Name:   connector.Status.Name,
			ID:     connector.Id.Id,
			Status: connector.Status.Connector.State,
			Type:   connector.Info.Type,
			Trace:  connector.Status.Connector.Trace,
		},
		Tasks:   []taskDescribeDisplay{},
		Configs: []configDescribeDisplay{},
	}
	for _, task := range connector.Status.Tasks {
		structuredDisplay.Tasks = append(structuredDisplay.Tasks, taskDescribeDisplay{
			TaskId: task.Id,
			State:  task.State,
		})
	}
	for name, value := range connector.Info.Config {
		structuredDisplay.Configs = append(structuredDisplay.Configs, configDescribeDisplay{
			Config: name,
			Value:  value,
		})
	}
	return output.StructuredOutput(format, structuredDisplay)
}
