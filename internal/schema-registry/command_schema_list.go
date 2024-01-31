package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type row struct {
	SchemaId int32  `human:"Schema ID" serialized:"schema_id"`
	Subject  string `human:"Subject" serialized:"subject"`
	Version  int32  `human:"Version" serialized:"version"`
}

func (c *command) newSchemaListCommand(cfg *config.Config) *cobra.Command {
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
		example1.Code += " " + onPremAuthenticationMsg
		example2.Code += " " + onPremAuthenticationMsg
		example3.Code += " " + onPremAuthenticationMsg
	}
	cmd.Example = examples.BuildExampleString(example1, example2, example3)

	cmd.Flags().String("subject-prefix", "", "List schemas for subjects with a given prefix.")
	cmd.Flags().Bool("all", false, "Include soft-deleted schemas.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		addCaLocationFlag(cmd)
		addSchemaRegistryEndpointFlag(cmd)
	}
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) schemaList(cmd *cobra.Command, _ []string) error {
	client, err := c.GetSchemaRegistryClient(cmd)
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

	schemas, err := client.GetSchemas(subjectPrefix, all)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, schema := range schemas {
		list.Add(&row{
			SchemaId: schema.GetId(),
			Subject:  schema.GetSubject(),
			Version:  schema.GetVersion(),
		})
	}
	return list.Print()
}
