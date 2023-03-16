package schemaregistry

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/properties"
	"github.com/confluentinc/cli/internal/pkg/resource"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
)

func (c *command) newExporterCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new schema exporter.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.exporterCreate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create a new schema exporter.",
				Code: fmt.Sprintf(`%s schema-registry exporter create my-exporter --config-file config.txt --subjects my-subject1,my-subject2 --subject-format my-\${subject} --context-type CUSTOM --context-name my-context`, pversion.CLIName),
			},
		),
	}

	cmd.Flags().String("config-file", "", "Exporter config file.")
	cmd.Flags().StringSlice("subjects", []string{"*"}, "A comma-separated list of exporter subjects.")
	cmd.Flags().String("subject-format", "${subject}", "Exporter subject rename format. The format string can contain ${subject}, which will be replaced with default subject name.")
	addContextTypeFlag(cmd)
	cmd.Flags().String("context-name", "", "Exporter context name.")
	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("config-file")

	return cmd
}

func (c *command) exporterCreate(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := getApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}

	return createExporter(cmd, args[0], srClient, ctx)
}

func createExporter(cmd *cobra.Command, name string, srClient *srsdk.APIClient, ctx context.Context) error {
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
		Name:                name,
		Subjects:            subjects,
		SubjectRenameFormat: subjectFormat,
		ContextType:         contextType,
		Context:             contextName,
		Config:              configMap,
	}

	if _, _, err := srClient.DefaultApi.CreateExporter(ctx, req); err != nil {
		return err
	}

	output.Printf(errors.CreatedResourceMsg, resource.SchemaExporter, name)
	return nil
}
