package schemaregistry

import (
	"context"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"
)

func (c *compatibilityCommand) newValidateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate input schema against a particular version of a subject for compatibility.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.validate),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Validate the compatibility of <schema> against a subject of latest version.",
				Code: fmt.Sprintf("%s schema-registry compatibility validate -S <subject-name> -V latest --schema <schema-path>", pversion.CLIName),
			},
		),
	}

	cmd.Flags().StringP("subject", "S", "", SubjectUsage)
	cmd.Flags().StringP("version", "V", "", "Version of the schema. Can be a specific version or 'latest'.")
	cmd.Flags().String("schema", "", "The path to the schema file.")
	cmd.Flags().String("type", "", `Specify the schema type as "avro", "protobuf", or "jsonschema".`)
	cmd.Flags().String("refs", "", "The path to the references file.")
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *compatibilityCommand) validate(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}

	return validateSchemaCompatibility(cmd, srClient, ctx)
}

func validateSchemaCompatibility(cmd *cobra.Command, srClient *srsdk.APIClient, ctx context.Context) error {
	subject, err := cmd.Flags().GetString("subject")
	if err != nil {
		return err
	}
	version, err := cmd.Flags().GetString("version")
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

	schema, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		return err
	}

	refs, err := readSchemaRefs(cmd)
	if err != nil {
		return err
	}

	req := srsdk.RegisterSchemaRequest{Schema: string(schema), SchemaType: schemaType, References: refs}

	compatibleResp, _, err := srClient.DefaultApi.TestCompatibilityBySubjectName(ctx, subject, version, req, nil)
	if err != nil {
		return err
	}
	compatible := strconv.FormatBool(compatibleResp.IsCompatible)

	outputOption, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	}
	if outputOption == output.Human.String() {
		printConfig(&struct{ Compatibility string }{compatible}, []string{"Compatibility"})
	} else {
		structuredOutput := &struct{ Compatibility string }{compatible}
		fields := []string{"Compatibility"}
		structuredRenames := map[string]string{"Compatibility": "compatibility"}
		return output.DescribeObject(cmd, structuredOutput, fields, map[string]string{}, structuredRenames)
	}
	return nil

}
