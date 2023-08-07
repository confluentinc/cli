package schemaregistry

import (
	"strings"

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

func (c *command) newExporterUpdateCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Update schema exporter.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.exporterUpdate,
	}

	example1 := examples.Example{
		Text: "Update schema exporter information.",
		Code: `confluent schema-registry exporter update my-exporter --subjects my-subject1,my-subject2 --subject-format my-\${subject} --context-type custom --context-name my-context`,
	}
	example2 := examples.Example{
		Text: "Update schema exporter configuration.",
		Code: "confluent schema-registry exporter update my-exporter --config-file config.txt",
	}
	if cfg.IsOnPremLogin() {
		example1.Code += " " + onPremAuthenticationMsg
		example2.Code += " " + onPremAuthenticationMsg
	}
	cmd.Example = examples.BuildExampleString(example1, example2)

	cmd.Flags().String("config-file", "", "Exporter configuration file.")
	cmd.Flags().StringSlice("subjects", []string{}, "A comma-separated list of exporter subjects.")
	cmd.Flags().String("subject-format", "${subject}", "Exporter subject rename format. The format string can contain ${subject}, which will be replaced with the default subject name.")
	addContextTypeFlag(cmd)
	cmd.Flags().String("context-name", "", "Exporter context name.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		addCaLocationFlag(cmd)
		addSchemaRegistryEndpointFlag(cmd)
	}
	pcmd.AddOutputFlag(cmd)

	if cfg.IsCloudLogin() {
		// Deprecated
		pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
		cobra.CheckErr(cmd.Flags().MarkHidden("api-key"))

		// Deprecated
		pcmd.AddApiSecretFlag(cmd)
		cobra.CheckErr(cmd.Flags().MarkHidden("api-secret"))
	}

	return cmd
}

func (c *command) exporterUpdate(cmd *cobra.Command, args []string) error {
	client, err := c.GetSchemaRegistryClient(cmd)
	if err != nil {
		return err
	}

	info, err := client.GetExporterInfo(args[0])
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
		updateRequest.ContextType = strings.ToUpper(contextType)
	}

	contextName, err := cmd.Flags().GetString("context-name")
	if err != nil {
		return err
	}
	if contextName != "" {
		updateRequest.Context = contextName
	}

	subjects, err := cmd.Flags().GetStringSlice("subjects")
	if err != nil {
		return err
	}
	if len(subjects) > 0 {
		updateRequest.Subjects = subjects
	}

	subjectFormat, err := cmd.Flags().GetString("subject-format")
	if err != nil {
		return err
	}
	if subjectFormat != "" {
		updateRequest.SubjectRenameFormat = subjectFormat
	}

	configFile, err := cmd.Flags().GetString("config-file")
	if err != nil {
		return err
	}

	if configFile != "" {
		updateRequest.Config, err = properties.FileToMap(configFile)
		if err != nil {
			return err
		}
	}

	if _, err := client.PutExporter(args[0], updateRequest); err != nil {
		return err
	}

	output.Printf(errors.UpdatedResourceMsg, resource.SchemaExporter, args[0])
	return nil
}
