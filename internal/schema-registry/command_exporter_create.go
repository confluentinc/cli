package schemaregistry

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/properties"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *command) newExporterCreateCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create <name>",
		Short:   "Create a new schema exporter.",
		Args:    cobra.ExactArgs(1),
		RunE:    c.exporterCreate,
		Example: examples.BuildExampleString(),
	}

	example := examples.Example{
		Text: "Create a new schema exporter.",
		Code: `confluent schema-registry exporter create my-exporter --config config.txt --subjects my-subject1,my-subject2 --subject-format my-\${subject} --context-type custom --context-name my-context`,
	}
	if cfg.IsOnPremLogin() {
		example.Code += " " + onPremAuthenticationMsg
	}
	cmd.Example = examples.BuildExampleString(example)

	pcmd.AddConfigFlag(cmd)
	cmd.Flags().StringSlice("subjects", []string{"*"}, "A comma-separated list of exporter subjects.")
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

	// Deprecated
	cmd.Flags().String("config-file", "", "Exporter configuration file.")
	cobra.CheckErr(cmd.Flags().MarkHidden("config-file"))
	cmd.MarkFlagsMutuallyExclusive("config", "config-file")

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

func (c *command) exporterCreate(cmd *cobra.Command, args []string) error {
	client, err := c.GetSchemaRegistryClient(cmd)
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
	contextType = strings.ToUpper(contextType)

	contextName := "."
	if contextType == "CUSTOM" {
		contextName, err = cmd.Flags().GetString("context-name")
		if err != nil {
			return err
		}
	} else if cmd.Flags().Changed("context-name") {
		return fmt.Errorf(`can only set context name if context type is "custom"`)
	}

	subjectFormat, err := cmd.Flags().GetString("subject-format")
	if err != nil {
		return err
	}

	config, err := cmd.Flags().GetStringSlice("config")
	if err != nil {
		return err
	}

	// Deprecated
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

	req := srsdk.CreateExporterRequest{
		Name:                srsdk.PtrString(args[0]),
		Subjects:            &subjects,
		SubjectRenameFormat: srsdk.PtrString(subjectFormat),
		ContextType:         srsdk.PtrString(contextType),
		Context:             srsdk.PtrString(contextName),
		Config:              &configMap,
	}

	if _, err := client.CreateExporter(req); err != nil {
		return err
	}

	output.Printf(c.Config.EnableColor, errors.CreatedResourceMsg, resource.SchemaExporter, args[0])
	return nil
}
