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

	cmd.Flags().String("log-level", "INFO", "Log level of custom code logging.")
	cmd.MarkFlagsOneRequired("log-level")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	return cmd
}

func (c *customCodeLoggingCommand) update(cmd *cobra.Command, args []string) error {
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

	if _, err := c.V2Client.UpdateCustomCodeLogging(id, updateCustomPluginRequest); err != nil {
		return err
	}

	output.Printf(c.Config.EnableColor, errors.UpdatedResourceMsg, resource.CustomCodeLogging, args[0])
	return nil
}
