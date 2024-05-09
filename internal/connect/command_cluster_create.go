package connect

import (
	"github.com/spf13/cobra"

	connectv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/kafka"
	"github.com/confluentinc/cli/v3/pkg/output"
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
				Text: "Create a configuration file with connector configs and offsets.",
				Code: `{
  "name": "MyGcsLogsBucketConnector",
  "config": {
    "connector.class": "GcsSink",
    "data.format": "BYTES",
    "flush.size": "1000",
    "gcs.bucket.name": "APILogsBucket",
    "gcs.credentials.config": "****************",
    "kafka.api.key": "****************",
    "kafka.api.secret": "****************",
    "name": "MyGcsLogsBucketConnector",
    "tasks.max": "2",
    "time.interval": "DAILY",
    "topics": "APILogsTopic"
  },
  "offsets": [
	{
	  "partition": {
		"kafka_partition": 0,
		"kafka_topic": "topic_A"
	  },
	  "offset": {
		"kafka_offset": 1000
	  }
	}
  ]
}`,
			},
			examples.Example{
				Text: "Create a connector in the current or specified Kafka cluster context.",
				Code: "confluent connect cluster create --config-file config.json",
			},
			examples.Example{
				Code: "confluent connect cluster create --config-file config.json --cluster lkc-123456",
			},
		),
	}

	cmd.Flags().String("config-file", "", "JSON connector configuration file.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagFilename("config-file", "json"))

	cobra.CheckErr(cmd.MarkFlagRequired("config-file"))

	return cmd
}

func (c *clusterCommand) create(cmd *cobra.Command, _ []string) error {
	kafkaCluster, err := kafka.GetClusterForCommand(c.V2Client, c.Context)
	if err != nil {
		return err
	}

	userConfigs, offsets, err := getConfigAndOffsets(cmd, false)
	if err != nil {
		return err
	}

	connectConfig := connectv1.InlineObject{
		Name:    connectv1.PtrString((userConfigs)["name"]),
		Config:  &userConfigs,
		Offsets: &offsets,
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	connectorInfo, err := c.V2Client.CreateConnector(environmentId, kafkaCluster.ID, connectConfig)
	if err != nil {
		return err
	}

	connector, err := c.V2Client.GetConnectorExpansionByName(connectorInfo.GetName(), environmentId, kafkaCluster.ID)
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
