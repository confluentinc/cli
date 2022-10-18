package schemaregistry

import (
	"context"
	"fmt"

	"github.com/antihax/optional"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"
)

var (
	fields           = []string{"SchemaID", "Subject", "Version"}
	humanLabels      = []string{"Schema ID", "Subject", "Version"}
	structuredLabels = []string{"schema_id", "subject", "version"}
)

type row struct {
	SchemaID int32
	Subject  string
	Version  int32
}

func (c *schemaCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List all schemas for given subject prefix.",
		Args:        cobra.NoArgs,
		RunE:        c.list,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List all schemas for given subject under default context.",
				Code: fmt.Sprintf("%s schema-registry schema list --subject-prefix my-subject", pversion.CLIName),
			},
			examples.Example{
				Text: "List all schemas under given context.",
				Code: fmt.Sprintf("%s schema-registry schema list --subject-prefix :.mycontext:", pversion.CLIName),
			},
			examples.Example{
				Text: "List all schemas under default context.",
				Code: fmt.Sprintf("%s schema-registry schema list", pversion.CLIName),
			},
		),
	}

	cmd.Flags().StringP("subject-prefix", "S", "", "Subject prefix to list the schemas from.")
	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *schemaCommand) list(cmd *cobra.Command, _ []string) error {
	srClient, ctx, err := getApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}

	return c.listSchemas(cmd, srClient, ctx)
}

func (c *schemaCommand) listSchemas(cmd *cobra.Command, srClient *srsdk.APIClient, ctx context.Context) error {
	subjectPrefix, err := cmd.Flags().GetString("subject-prefix")
	if err != nil {
		return err
	}

	getSchemasOpts := srsdk.GetSchemasOpts{SubjectPrefix: optional.NewString(subjectPrefix)}
	schemas, _, err := srClient.DefaultApi.GetSchemas(ctx, &getSchemasOpts)
	if err != nil {
		return err
	}

	outputWriter, err := output.NewListOutputWriter(cmd, fields, humanLabels, structuredLabels)
	if err != nil {
		return err
	}
	for _, schema := range schemas {
		outputWriter.AddElement(&row{
			SchemaID: schema.Id,
			Subject:  schema.Subject,
			Version:  schema.Version,
		})
	}
	return outputWriter.Out()
}
