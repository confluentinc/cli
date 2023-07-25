package schemaregistry

import (
	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/properties"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *command) newExporterCreateCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create <name>",
		Short:   "Create a new schema exporter.",
		Args:    cobra.ExactArgs(1),
		RunE:    c.exporterCreate,
		Example: examples.BuildExampleString(),
	}

	example := examples.Example{
		Text: "Create a new schema exporter.",
		Code: `confluent schema-registry exporter create my-exporter --config-file config.txt --subjects my-subject1,my-subject2 --subject-format my-\${subject} --context-type CUSTOM --context-name my-context`,
	}
	if !cfg.IsCloudLogin() {
		example.Code += " " + OnPremAuthenticationMsg
	}
	cmd.Example = examples.BuildExampleString(example)

	cmd.Flags().String("config-file", "", "Exporter configuration file.")
	cmd.Flags().StringSlice("subjects", []string{"*"}, "A comma-separated list of exporter subjects.")
	cmd.Flags().String("subject-format", "${subject}", "Exporter subject rename format. The format string can contain ${subject}, which will be replaced with the default subject name.")
	addContextTypeFlag(cmd)
	cmd.Flags().String("context-name", "", "Exporter context name.")
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		cmd.Flags().AddFlagSet(pcmd.OnPremSchemaRegistrySet())
	}
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("config-file"))

	return cmd
}

func (c *command) exporterCreate(cmd *cobra.Command, args []string) error {
	client, err := c.GetSchemaRegistryClient()
	if err != nil {
		return err
	}

	subjects, err := cmd.Flags().GetStringSlice("subjects")
	if err != nil {
		return err
	}

	contextType, err := cmd.Flags().GetString("context-type")
	if err != nil {
		return err
	}

	contextName := "."
	if contextType == "CUSTOM" {
		contextName, err = cmd.Flags().GetString("context-name")
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

	configMap := make(map[string]string)
	if configFile != "" {
		configMap, err = properties.FileToMap(configFile)
		if err != nil {
			return err
		}
	}

	req := srsdk.CreateExporterRequest{
		Name:                args[0],
		Subjects:            subjects,
		SubjectRenameFormat: subjectFormat,
		ContextType:         contextType,
		Context:             contextName,
		Config:              configMap,
	}

	if _, err := client.CreateExporter(req); err != nil {
		return err
	}

	output.Printf(errors.CreatedResourceMsg, resource.SchemaExporter, args[0])
	return nil
}
