package customcodelogging

import (
	"github.com/spf13/cobra"

	cclv1 "github.com/confluentinc/ccloud-sdk-go-v2/ccl/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *customCodeLoggingCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a custom code logging.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.update,
	}

	cmd.Flags().String("destination-topic", "", "Kafka topic of custom code logging destination.")
	cmd.Flags().String("destination-cluster-id", "", "Kafka cluster id of custom code logging destination.")
	cmd.Flags().String("log-level", "", "Log level of custom code logging. (default \"INFO\")")
	cmd.MarkFlagsOneRequired("destination-topic", "destination-cluster-id", "log-level")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	return cmd
}

func (c *customCodeLoggingCommand) update(cmd *cobra.Command, args []string) error {
	id := args[0]
	updateCustomPluginRequest := cclv1.CclV1CustomCodeLoggingUpdate{}

	if cmd.Flags().Changed("destination-topic") {
		if topic, err := cmd.Flags().GetString("destination-topic"); err != nil {
			return err
		} else {
			updateCustomPluginRequest.SetDestinationSettings(cclv1.CclV1CustomCodeLoggingUpdateDestinationSettingsOneOf{
				CclV1KafkaDestinationSettings: &cclv1.CclV1KafkaDestinationSettings{
					Topic: topic,
				},
			})
		}
	}

	if cmd.Flags().Changed("destination-cluster-id") {
		if clusterId, err := cmd.Flags().GetString("destination-cluster-id"); err != nil {
			return err
		} else {
			updateCustomPluginRequest.SetDestinationSettings(cclv1.CclV1CustomCodeLoggingUpdateDestinationSettingsOneOf{
				CclV1KafkaDestinationSettings: &cclv1.CclV1KafkaDestinationSettings{
					ClusterId: clusterId,
				},
			})
		}
	}

	if cmd.Flags().Changed("log-level") {
		if logLevel, err := cmd.Flags().GetString("log-level"); err != nil {
			return err
		} else {
			updateCustomPluginRequest.SetDestinationSettings(cclv1.CclV1CustomCodeLoggingUpdateDestinationSettingsOneOf{
				CclV1KafkaDestinationSettings: &cclv1.CclV1KafkaDestinationSettings{
					LogLevel: cclv1.PtrString(logLevel),
				},
			})
		}
	}

	if _, err := c.V2Client.UpdateCustomCodeLogging(id, updateCustomPluginRequest); err != nil {
		return err
	}

	output.Printf(c.Config.EnableColor, errors.UpdatedResourceMsg, resource.CustomCodeLogging, args[0])
	return nil
}
