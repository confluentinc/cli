package schemaregistry

import (
	"fmt"
	"strings"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
	"github.com/spf13/cobra"
)

func (c *schemaCommand) newCreateCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a schema.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.onPremCreate),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Register a new schema.",
				Code: fmt.Sprintf("%s schema-registry schema create --subject payments --schema schemafilepath %s", pversion.CLIName, errors.OnPremAuthenticationMsg),
			},
		),
	}

	cmd.Flags().String("schema", "", "The path to the schema file.")
	cmd.Flags().StringP("subject", "S", "", SubjectUsage)
	cmd.Flags().String("type", "", `Specify the schema type as "avro", "protobuf", or "jsonschema".`)
	cmd.Flags().String("refs", "", "The path to the references file.")
	cmd.Flags().String("sr-endpoint", "", "The URL of the schema registry cluster.")
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("schema")
	_ = cmd.MarkFlagRequired("subject")

	pcmd.RegisterFlagCompletionFunc(cmd, "type", func(_ *cobra.Command, _ []string) []string { return []string{"AVRO", "PROTOBUF", "JSON"} })

	return cmd
}

func (c *schemaCommand) onPremCreate(cmd *cobra.Command, _ []string) error {
	subject, err := cmd.Flags().GetString("subject")
	if err != nil {
		return err
	}
	schemaType, err := cmd.Flags().GetString("type")
	if err != nil {
		return err
	}
	schemaPath, err := cmd.Flags().GetString("schema")
	if err != nil {
		return err
	}
	refs, err := readSchemaRefs(cmd)
	if err != nil {
		return err
	}
	_, _, err = c.registerSchema(cmd, schemaType, schemaPath, subject, strings.ToUpper(schemaType), refs)
	if err != nil {
		return err
	}

	return nil
}
