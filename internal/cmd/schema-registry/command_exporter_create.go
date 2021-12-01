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

func (c *exporterCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create new schema exporter.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.create),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create a new schema exporter.",
				Code: `confluent schema-registry exporter create my-exporter --config-file config.txt --subjects my-subject1,my-subject2 --subject-format my-\${subject} --context-type CUSTOM --context-name my-context`,
			},
		),
	}

	cmd.Flags().String("config-file", "", "Exporter config file.")
	cmd.Flags().StringSlice("subjects", []string{"*"}, "Exporter subjects. Use a comma separated list, or specify the flag multiple times.")
	cmd.Flags().String("subject-format", "${subject}", "Exporter subject rename format. The format string can contain ${subject}, which will be replaced with default subject name.")
	cmd.Flags().String("context-type", "AUTO", `Exporter context type. One of "AUTO", "CUSTOM" or "NONE".`)
	cmd.Flags().String("context-name", "", "Exporter context name.")
	cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)

	_ = cmd.MarkFlagRequired("config-file")

	return cmd
}

func (c *exporterCommand) create(cmd *cobra.Command, args []string) error {
	name := args[0]

	subjects, err := cmd.Flags().GetStringSlice("subjects")
	if err != nil {
		return err
	}

	contextType, err := cmd.Flags().GetString("context-type")
	if err != nil {
		return err
	}

	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}

	context := "."
	if contextType == "CUSTOM" {
		context, err = cmd.Flags().GetString("context-name")
		if err != nil {
			return err
		}
	} else if cmd.Flags().Changed("context-name") {
		return errors.New("can only set context-name if context-type is CUSTOM")
	}

	subjectFormat, err := cmd.Flags().GetString("subject-format")
	if err != nil {
		return err
	}

	configFile, err := cmd.Flags().GetString("config-file")
	if err != nil {
		return err
	}

	configMap, err := utils.ReadConfigsFromFile(configFile)
	if err != nil {
		return err
	}

	req := srsdk.CreateExporterRequest{
		Name:                name,
		Subjects:            subjects,
		SubjectRenameFormat: subjectFormat,
		ContextType:         contextType,
		Context:             context,
		Config:              configMap,
	}

	if _, _, err = srClient.DefaultApi.CreateExporter(ctx, req); err != nil {
		return err
	}

	utils.Printf(cmd, errors.ExporterActionMsg, "Created", name)
	return nil
}
