package schemaregistry

import (
	"context"
	"fmt"

	"github.com/antihax/optional"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
)

var (
	fields           = []string{"SchemaId", "Subject", "Version"}
	humanLabels      = []string{"Schema ID", "Subject", "Version"}
	structuredLabels = []string{"schema_id", "subject", "version"}
)

type row struct {
	SchemaId int32
	Subject  string
	Version  int32
}

func (c *schemaCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List schemas for a given subject prefix.",
		Args:        cobra.NoArgs,
		RunE:        c.list,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `List all schemas for subjects with prefix "my-subject".`,
				Code: fmt.Sprintf("%s schema-registry schema list --subject-prefix my-subject", pversion.CLIName),
			},
			examples.Example{
				Text: `List all schemas for all subjects in context ":.mycontext:".`,
				Code: fmt.Sprintf("%s schema-registry schema list --subject-prefix :.mycontext:", pversion.CLIName),
			},
			examples.Example{
				Text: "List all schemas in default context.",
				Code: fmt.Sprintf("%s schema-registry schema list", pversion.CLIName),
			},
		),
	}

	cmd.Flags().String("subject-prefix", "", "List schemas for subjects with a given prefix.")
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

	opts := &srsdk.GetSchemasOpts{SubjectPrefix: optional.NewString(subjectPrefix)}
	schemas, _, err := srClient.DefaultApi.GetSchemas(ctx, opts)
	if err != nil {
		return err
	}

	outputWriter, err := output.NewListOutputWriter(cmd, fields, humanLabels, structuredLabels)
	if err != nil {
		return err
	}
	for _, schema := range schemas {
		outputWriter.AddElement(&row{
			SchemaId: schema.Id,
			Subject:  schema.Subject,
			Version:  schema.Version,
		})
	}
	return outputWriter.Out()
}
