package customcodelogging

import (
	"fmt"

	"github.com/spf13/cobra"

	cclv1 "github.com/confluentinc/ccloud-sdk-go-v2/ccl/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/resource"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

func (c *customCodeLoggingCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a custom code logging.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.update,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe custom code logging.",
				Code: "confluent custom-code-logging update ccl-123456 --log-level DEBUG --environment env-000000",
			},
		),
	}
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String("log-level", "INFO", fmt.Sprintf("Specify the Custom Code Logging Log Level as %s.", utils.ArrayToCommaDelimitedString(allowedLogLevels, "or")))
	cmd.MarkFlagsOneRequired("log-level")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	return cmd
}

func (c *customCodeLoggingCommand) update(cmd *cobra.Command, args []string) error {
	environment, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}
	id := args[0]
	updateCustomPluginRequest := cclv1.CclV1CustomCodeLoggingUpdate{}

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

	if _, err := c.V2Client.UpdateCustomCodeLogging(id, environment, updateCustomPluginRequest); err != nil {
		return err
	}

	output.Printf(c.Config.EnableColor, errors.UpdatedResourceMsg, resource.CustomCodeLogging, args[0])
	return nil
}
