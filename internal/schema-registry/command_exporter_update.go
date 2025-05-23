package schemaregistry

import (
	"strings"

	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/properties"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *command) newExporterUpdateCommand(cfg *config.Config) *cobra.Command {
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
		Code: "confluent schema-registry exporter update my-exporter --config config.txt",
	}
	if cfg.IsOnPremLogin() {
		example1.Code += " " + onPremAuthenticationMsg
		example2.Code += " " + onPremAuthenticationMsg
	}
	cmd.Example = examples.BuildExampleString(example1, example2)

	pcmd.AddConfigFlag(cmd)
	cmd.Flags().StringSlice("subjects", []string{}, "A comma-separated list of exporter subjects.")
	cmd.Flags().String("subject-format", "${subject}", "Exporter subject rename format. The format string can contain ${subject}, which will be replaced with the default subject name.")
	addContextTypeFlag(cmd)
	cmd.Flags().String("context-name", "", "Exporter context name.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		addCaLocationAndClientPathFlags(cmd)
	}
	addSchemaRegistryEndpointFlag(cmd)
	pcmd.AddOutputFlag(cmd)

	// Deprecated
	cmd.Flags().String("config-file", "", "Exporter configuration file.")
	cobra.CheckErr(cmd.Flags().MarkHidden("config-file"))
	cmd.MarkFlagsMutuallyExclusive("config", "config-file")

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
		updateRequest.ContextType = srsdk.PtrString(strings.ToUpper(contextType))
	}

	contextName, err := cmd.Flags().GetString("context-name")
	if err != nil {
		return err
	}
	if contextName != "" {
		updateRequest.Context = srsdk.PtrString(contextName)
	}

	subjects, err := cmd.Flags().GetStringSlice("subjects")
	if err != nil {
		return err
	}
	if len(subjects) > 0 {
		updateRequest.Subjects = &subjects
	}

	subjectFormat, err := cmd.Flags().GetString("subject-format")
	if err != nil {
		return err
	}
	if subjectFormat != "" {
		updateRequest.SubjectRenameFormat = srsdk.PtrString(subjectFormat)
	}

	config, err := cmd.Flags().GetStringSlice("config")
	if err != nil {
		return err
	}

	configFile, err := cmd.Flags().GetString("config-file")
	if err != nil {
		return err
	}
	if configFile != "" {
		config = []string{configFile}
	}

	configMap, err := properties.GetMap(config)
	if err != nil {
		return err
	}
	if len(configMap) > 0 {
		updateRequest.Config = &configMap
	}

	if _, err := client.PutExporter(args[0], updateRequest); err != nil {
		return err
	}

	output.Printf(c.Config.EnableColor, errors.UpdatedResourceMsg, resource.SchemaExporter, args[0])
	return nil
}
