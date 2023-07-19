package schemaregistry

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
)

func (c *command) newSchemaCreateCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "create",
		Short:       "Create a schema.",
		Args:        cobra.NoArgs,
		RunE:        c.schemaCreateOnPrem,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Register a new schema.",
				Code: fmt.Sprintf("confluent schema-registry schema create --subject payments --schema payments.avro --type avro %s", OnPremAuthenticationMsg),
			},
		),
	}

	cmd.Flags().String("schema", "", "The path to the schema file.")
	cmd.Flags().String("subject", "", SubjectUsage)
	pcmd.AddSchemaTypeFlag(cmd)
	cmd.Flags().String("references", "", "The path to the references file.")
	cmd.Flags().String("metadata", "", "The path to metadata file.")
	cmd.Flags().String("ruleset", "", "The path to schema ruleset file.")
	cmd.Flags().Bool("normalize", false, "Alphabetize the list of schema fields.")
	cmd.Flags().AddFlagSet(pcmd.OnPremSchemaRegistrySet())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagFilename("schema", "avsc", "json", "proto"))
	cobra.CheckErr(cmd.MarkFlagFilename("references", "json"))
	cobra.CheckErr(cmd.MarkFlagFilename("metadata", "json"))
	cobra.CheckErr(cmd.MarkFlagFilename("ruleset", "json"))

	cobra.CheckErr(cmd.MarkFlagRequired("schema"))
	cobra.CheckErr(cmd.MarkFlagRequired("subject"))

	return cmd
}

func (c *command) schemaCreateOnPrem(cmd *cobra.Command, _ []string) error {
	subject, err := cmd.Flags().GetString("subject")
	if err != nil {
		return err
	}

	schemaPath, err := cmd.Flags().GetString("schema")
	if err != nil {
		return err
	}

	schemaType, err := cmd.Flags().GetString("type")
	if err != nil {
		return err
	}
	schemaType = strings.ToUpper(schemaType)

	refs, err := ReadSchemaRefs(cmd)
	if err != nil {
		return err
	}

	dir, err := CreateTempDir()
	if err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	normalize, err := cmd.Flags().GetBool("normalize")
	if err != nil {
		return err
	}

	schemaCfg := &RegisterSchemaConfigs{
		SchemaDir:  dir,
		SchemaType: schemaType,
		SchemaPath: schemaPath,
		Subject:    subject,
		Refs:       refs,
		Normalize:  normalize,
	}

	metadata, err := cmd.Flags().GetString("metadata")
	if err != nil {
		return err
	}
	if metadata != "" {
		schemaCfg.Metadata = new(srsdk.Metadata)
		if err := read(metadata, schemaCfg.Metadata); err != nil {
			return err
		}
	}

	ruleset, err := cmd.Flags().GetString("ruleset")
	if err != nil {
		return err
	}
	if ruleset != "" {
		schemaCfg.Ruleset = new(srsdk.RuleSet)
		if err := read(ruleset, schemaCfg.Ruleset); err != nil {
			return err
		}
	}

	return c.registerSchemaOnPrem(cmd, schemaCfg)
}

func (c *command) registerSchemaOnPrem(cmd *cobra.Command, schemaCfg *RegisterSchemaConfigs) error {
	if c.State == nil { // require log-in to use oauthbearer token
		return errors.NewErrorWithSuggestions(errors.NotLoggedInErrorMsg, errors.AuthTokenSuggestions)
	}

	srClient, ctx, err := GetSrApiClientWithToken(cmd, c.Version, c.AuthToken())
	if err != nil {
		return err
	}

	if _, err := RegisterSchemaWithAuth(cmd, schemaCfg, srClient, ctx); err != nil {
		return err
	}

	return nil
}
