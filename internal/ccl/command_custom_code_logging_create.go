package customcodelogging

import (
	"github.com/confluentinc/cli/v4/pkg/output"
	"strings"

	cclv1 "github.com/confluentinc/ccloud-sdk-go-v2/ccl/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
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
				Code: "confluent ccl custom-code-logging create --cloud aws --region us-west-2 --environment env-000000 --destination-kafka --destination-topic topic-123 --destination-cluster-id cluster-123",
			},
		),
	}

	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlagKafka(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().Bool("destination-kafka", true, "Set custom code logging destination to KAFKA")
	cmd.Flags().String("destination-topic", "", "Kafka topic of custom code logging destination.")
	cmd.Flags().String("destination-cluster-id", "", "Kafka cluster id of custom code logging destination.")
	cmd.Flags().String("log-level", "", "Log level of custom code logging. (default \"INFO\")")

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("region"))
	cobra.CheckErr(cmd.MarkFlagRequired("environment"))
	cmd.MarkFlagsOneRequired("destination-kafka")
	cmd.MarkFlagsRequiredTogether("destination-kafka", "destination-topic", "destination-cluster-id")
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

	logLevel, err := cmd.Flags().GetString("log-level")

	destinationKafka, err := cmd.Flags().GetBool("destination-kafka")
	if err != nil {
		return err
	}

	topic, err := cmd.Flags().GetString("destination-topic")
	if err != nil {
		return err
	}

	clusterId, err := cmd.Flags().GetString("destination-cluster-id")
	if err != nil {
		return err
	}

	var customCodeLoggingDestinationSettings *cclv1.CclV1CustomCodeLoggingDestinationSettingsOneOf
	if destinationKafka {
		customCodeLoggingDestinationSettings = &cclv1.CclV1CustomCodeLoggingDestinationSettingsOneOf{
			CclV1KafkaDestinationSettings: &cclv1.CclV1KafkaDestinationSettings{
				Kind:      "KAFKA",
				Topic:     topic,
				ClusterId: clusterId,
			},
		}
		if logLevel != "" {
			customCodeLoggingDestinationSettings.CclV1KafkaDestinationSettings.SetLogLevel(logLevel)
		}
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
