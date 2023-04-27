package schemaregistry

import (
	"context"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type row struct {
	SchemaId int32  `human:"Schema ID" serialized:"schema_id"`
	Subject  string `human:"Subject" serialized:"subject"`
	Version  int32  `human:"Version" serialized:"version"`
}

func (c *command) newSchemaListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List schemas for a given subject prefix.",
		Args:        cobra.NoArgs,
		RunE:        c.schemaList,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `List all schemas for subjects with prefix "my-subject".`,
				Code: "confluent schema-registry schema list --subject-prefix my-subject",
			},
			examples.Example{
				Text: `List all schemas for all subjects in context ":.mycontext:".`,
				Code: "confluent schema-registry schema list --subject-prefix :.mycontext:",
			},
			examples.Example{
				Text: "List all schemas in the default context.",
				Code: "confluent schema-registry schema list",
			},
		),
	}

	cmd.Flags().String("subject-prefix", "", "List schemas for subjects with a given prefix.")
	cmd.Flags().Bool("all", false, "Include soft-deleted schemas.")
	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) schemaList(cmd *cobra.Command, _ []string) error {
	srClient, ctx, err := getApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}

	return c.listSchemas(cmd, srClient, ctx)
}

func (c *command) listSchemas(cmd *cobra.Command, srClient *srsdk.APIClient, ctx context.Context) error {
	subjectPrefix, err := cmd.Flags().GetString("subject-prefix")
	if err != nil {
		return err
	}

	all, err := cmd.Flags().GetBool("all")
	if err != nil {
		return err
	}

	opts := &srsdk.GetSchemasOpts{SubjectPrefix: optional.NewString(subjectPrefix), Deleted: optional.NewBool(showDeleted)}
	schemas, _, err := srClient.DefaultApi.GetSchemas(ctx, opts)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, schema := range schemas {
		list.Add(&row{
			SchemaId: schema.Id,
			Subject:  schema.Subject,
			Version:  schema.Version,
		})
	}
	return list.Print()
}
