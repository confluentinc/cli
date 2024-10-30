package customcodelogging

import (
	"strings"

	"github.com/spf13/cobra"

	cclv1 "github.com/confluentinc/ccloud-sdk-go-v2/ccl/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *customCodeLoggingCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a custom code logging.",
		Args:  cobra.ExactArgs(0),
		RunE:  c.createCustomCodeLogging,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create custom code logging.`,
				Code: "confluent custom-code-logging create --cloud aws --region us-west-2 --environment env-000000 --topic topic-123 --cluster-id cluster-123",
			},
		),
	}

	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlagKafka(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String("topic", "", "Kafka topic of custom code logging destination.")
	cmd.Flags().String("cluster-id", "", "Kafka cluster id of custom code logging destination.")
	cmd.Flags().String("log-level", "", "Log level of custom code logging. (default \"INFO\").")

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("region"))
	cobra.CheckErr(cmd.MarkFlagRequired("environment"))
	cmd.MarkFlagsOneRequired("topic")
	cmd.MarkFlagsOneRequired("cluster-id")
	return cmd
}

func (c *customCodeLoggingCommand) createCustomCodeLogging(cmd *cobra.Command, args []string) error {
	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}

	cloud = strings.ToUpper(cloud)

	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return err
	}

	region = strings.ToUpper(region)

	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	logLevel, _ := cmd.Flags().GetString("log-level")

	topic, err := cmd.Flags().GetString("topic")
	if err != nil {
		return err
	}

	clusterId, err := cmd.Flags().GetString("cluster-id")
	if err != nil {
		return err
	}

	var customCodeLoggingDestinationSettings = &cclv1.CclV1CustomCodeLoggingDestinationSettingsOneOf{
		CclV1KafkaDestinationSettings: &cclv1.CclV1KafkaDestinationSettings{
			Kind:      "Kafka",
			Topic:     topic,
			ClusterId: clusterId,
		},
	}
	if logLevel != "" {
		customCodeLoggingDestinationSettings.CclV1KafkaDestinationSettings.SetLogLevel(logLevel)
	}

	request := cclv1.CclV1CustomCodeLogging{
		Cloud:               cclv1.PtrString(cloud),
		Region:              cclv1.PtrString(region),
		Environment:         &cclv1.EnvScopedObjectReference{Id: environment},
		DestinationSettings: customCodeLoggingDestinationSettings,
	}

	resp, err := c.V2Client.CreateCustomCodeLogging(request)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&customCodeLoggingShortOut{
		Id:          resp.GetId(),
		Cloud:       resp.GetCloud(),
		Region:      resp.GetRegion(),
		Environment: resp.GetEnvironment().Id,
	})
	return table.Print()
}
