package schemaregistry

import (
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *exporterCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Update configs or information of schema exporter.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.update),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Update information of new schema exporter.",
				Code: "confluent schema-registry exporter update my-exporter --subjects my-subject1,my-subject2 --context-type CUSTOM --context-name my-context",
			},
			examples.Example{
				Text: "Update configs of new schema exporter.",
				Code: "confluent schema-registry exporter update my-exporter --config-file ~/config.txt",
			},
		),
	}

	cmd.Flags().String("config-file", "", "The file containing configurations of the exporter.")
	cmd.Flags().StringSlice("subjects", []string{}, "The subjects of the exporter. Should use comma separated list, or specify the flag multiple times.")
	cmd.Flags().String("context-type", "", `The context type of the exporter. Can be "AUTO", "CUSTOM" or "NONE".`)
	cmd.Flags().String("context-name", "", "The context name of the exporter.")
	cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)

	return cmd
}

func (c *exporterCommand) update(cmd *cobra.Command, args []string) error {
	name := args[0]

	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}

	info, _, err := srClient.DefaultApi.GetExporterInfo(ctx, name)
	if err != nil {
		return err
	}

	updateRequest := srsdk.UpdateExporterRequest{
		Subjects:    info.Subjects,
		ContextType: info.ContextType,
		Context:     info.Context,
	}

	contextType, err := cmd.Flags().GetString("context-type")
	if err != nil {
		return err
	}
	if contextType != "" {
		updateRequest.ContextType = contextType
	}

	context, err := cmd.Flags().GetString("context-name")
	if err != nil {
		return err
	}
	if context != "" {
		updateRequest.Context = context
	}

	subjects, err := cmd.Flags().GetStringSlice("subjects")
	if err != nil {
		return err
	}
	if len(subjects) > 0 {
		updateRequest.Subjects = subjects
	}

	configFile, err := cmd.Flags().GetString("config-file")
	if err != nil {
		return err
	}
	if configFile != "" {
		configMap, err := utils.ReadConfigsFromFile(configFile)
		if err != nil {
			return err
		}
		updateRequest.Config = configMap
	}

	if _, _, err = srClient.DefaultApi.PutExporter(ctx, name, updateRequest); err != nil {
		return err
	}

	utils.Printf(cmd, errors.ExporterActionMsg, "Updated", name)
	return nil
}
