package connect

import (
	"github.com/spf13/cobra"

	connectv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect/v1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type connectCreateOut struct {
	Id         string `human:"ID" serialized:"id"`
	Name       string `human:"Name" serialized:"name"`
	ErrorTrace string `human:"Error Trace,omitempty" serialized:"error_trace,omitempty"`
}

func (c *clusterCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "create",
		Short:       "Create a connector.",
		Args:        cobra.NoArgs,
		RunE:        c.create,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create a connector in the current or specified Kafka cluster context.",
				Code: "confluent connect cluster create --config-file config.json",
			},
			examples.Example{
				Code: "confluent connect cluster create --config-file config.json --cluster lkc-123456",
			},
		),
	}

	cmd.Flags().String("config-file", "", "JSON connector config file.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagFilename("config-file", "json"))

	cobra.CheckErr(cmd.MarkFlagRequired("config-file"))

	return cmd
}

func (c *clusterCommand) create(cmd *cobra.Command, _ []string) error {
	kafkaCluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	userConfigs, err := getConfig(cmd)
	if err != nil {
		return err
	}

	connectConfig := connectv1.InlineObject{
		Name:   connectv1.PtrString((*userConfigs)["name"]),
		Config: userConfigs,
	}

	connectorInfo, err := c.V2Client.CreateConnector(c.EnvironmentId(), kafkaCluster.ID, connectConfig)
	if err != nil {
		return err
	}

	connector, err := c.V2Client.GetConnectorExpansionByName(connectorInfo.GetName(), c.EnvironmentId(), kafkaCluster.ID)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&connectCreateOut{
		Id:         connector.Id.GetId(),
		Name:       connectorInfo.GetName(),
		ErrorTrace: connector.Status.Connector.GetTrace(),
	})
	return table.Print()
}
