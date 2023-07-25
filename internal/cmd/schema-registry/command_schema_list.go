package schemaregistry

import (
	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type row struct {
	SchemaId int32  `human:"Schema ID" serialized:"schema_id"`
	Subject  string `human:"Subject" serialized:"subject"`
	Version  int32  `human:"Version" serialized:"version"`
}

func (c *command) newSchemaListCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List schemas for a given subject prefix.",
		Args:  cobra.NoArgs,
		RunE:  c.schemaList,
	}

	example1 := examples.Example{
		Text: `List all schemas for subjects with prefix "my-subject".`,
		Code: "confluent schema-registry schema list --subject-prefix my-subject",
	}
	example2 := examples.Example{
		Text: `List all schemas for all subjects in context ":.mycontext:".`,
		Code: "confluent schema-registry schema list --subject-prefix :.mycontext:",
	}
	example3 := examples.Example{
		Text: "List all schemas in the default context.",
		Code: "confluent schema-registry schema list",
	}
	if cfg.IsOnPremLogin() {
		example1.Code += " " + OnPremAuthenticationMsg
		example2.Code += " " + OnPremAuthenticationMsg
		example3.Code += " " + OnPremAuthenticationMsg
	}
	cmd.Example = examples.BuildExampleString(example1, example2, example3)

	cmd.Flags().String("subject-prefix", "", "List schemas for subjects with a given prefix.")
	cmd.Flags().Bool("all", false, "Include soft-deleted schemas.")
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		cmd.Flags().AddFlagSet(pcmd.OnPremSchemaRegistrySet())
	}
	pcmd.AddContextFlag(cmd, c.CLICommand)
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

func (c *command) schemaList(cmd *cobra.Command, _ []string) error {
	client, err := c.GetSchemaRegistryClient()
	if err != nil {
		return err
	}

	subjectPrefix, err := cmd.Flags().GetString("subject-prefix")
	if err != nil {
		return err
	}

	all, err := cmd.Flags().GetBool("all")
	if err != nil {
		return err
	}

	opts := &srsdk.GetSchemasOpts{SubjectPrefix: optional.NewString(subjectPrefix), Deleted: optional.NewBool(all)}
	schemas, err := client.GetSchemas(opts)
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
