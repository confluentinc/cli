package customcodelogging

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	cclv1 "github.com/confluentinc/ccloud-sdk-go-v2/ccl/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

var (
	allowedLogLevels = []string{"INFO", "DEBUG", "ERROR", "WARN"}
)

func (c *customCodeLoggingCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a custom code logging.",
		Args:  cobra.NoArgs,
		RunE:  c.createCustomCodeLogging,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create custom code logging.`,
				Code: "confluent custom-code-logging create --cloud aws --region us-west-2 --topic topic-123 --cluster cluster-123 --environment env-000000",
			},
		),
	}

	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlagKafka(cmd, c.AuthenticatedCLICommand)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String("topic", "", "Kafka topic of custom code logging destination.")
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String("log-level", "INFO", fmt.Sprintf("Specify the Custom Code Logging Log Level as %s.", utils.ArrayToCommaDelimitedString(allowedLogLevels, "or")))
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("region"))
	cobra.CheckErr(cmd.MarkFlagRequired("topic"))
	cobra.CheckErr(cmd.MarkFlagRequired("cluster"))
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

	environment, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	logLevel, err := cmd.Flags().GetString("log-level")
	if err != nil {
		return err
	}

	topic, err := cmd.Flags().GetString("topic")
	if err != nil {
		return err
	}

	clusterId, err := cmd.Flags().GetString("cluster")
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
	table.Add(getCustomCodeLogging(resp))
	table.Filter([]string{"Id", "Cloud", "Region", "Environment", "Topic", "Cluster", "LogLevel"})
	return table.Print()
}
