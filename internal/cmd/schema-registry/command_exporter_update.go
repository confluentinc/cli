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

func (c *command) newExporterUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Update configs or information of schema exporter.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.exporterUpdate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Update information of new schema exporter.",
				Code: fmt.Sprintf(`%s schema-registry exporter update my-exporter --subjects my-subject1,my-subject2 --subject-format my-\${subject} --context-type CUSTOM --context-name my-context`, pversion.CLIName),
			},
			examples.Example{
				Text: "Update configs of new schema exporter.",
				Code: fmt.Sprintf(`%s schema-registry exporter update my-exporter --config-file ~/config.txt`, pversion.CLIName),
			},
		),
	}

	cmd.Flags().String("config-file", "", "Exporter configuration file.")
	cmd.Flags().StringSlice("subjects", []string{}, "A comma-separated list of exporter subjects.")
	cmd.Flags().String("subject-format", "${subject}", "Exporter subject rename format. The format string can contain ${subject}, which will be replaced with the default subject name.")
	addContextTypeFlag(cmd)
	cmd.Flags().String("context-name", "", "Exporter context name.")
	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) exporterUpdate(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := getApiClient(cmd, c.Config, c.Version)
	if err != nil {
		return err
	}

	return updateExporter(cmd, args[0], srClient, ctx)
}

func updateExporter(cmd *cobra.Command, name string, srClient *srsdk.APIClient, ctx context.Context) error {
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

	if _, _, err := srClient.DefaultApi.PutExporter(ctx, name, updateRequest); err != nil {
		return err
	}

	output.Printf(errors.UpdatedResourceMsg, resource.SchemaExporter, name)
	return nil
}
