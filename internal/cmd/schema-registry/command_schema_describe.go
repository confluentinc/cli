package schemaregistry

import (
	"fmt"
	"strconv"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/utils"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
)

func (c *schemaCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "describe <id>",
		Short:   "Get schema either by schema ID, or by subject/version.",
		Args:    cobra.MaximumNArgs(1),
		PreRunE: pcmd.NewCLIPreRunnerE(c.preDescribe),
		RunE:    pcmd.NewCLIRunE(c.describe),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe the schema string by schema ID.",
				Code: fmt.Sprintf("%s schema-registry schema describe 1337", pversion.CLIName),
			},
			examples.Example{
				Text: "Describe the schema by both subject and version.",
				Code: fmt.Sprintf("%s schema-registry schema describe --subject payments --version latest", pversion.CLIName),
			},
		),
	}

	cmd.Flags().StringP("subject", "S", "", SubjectUsage)
	cmd.Flags().StringP("version", "V", "", "Version of the schema. Can be a specific version or 'latest'.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *schemaCommand) preDescribe(cmd *cobra.Command, args []string) error {
	subject, err := cmd.Flags().GetString("subject")
	if err != nil {
		return err
	}

	version, err := cmd.Flags().GetString("version")
	if err != nil {
		return err
	}

	if len(args) > 0 && (subject != "" || version != "") {
		return errors.New(errors.BothSchemaAndSubjectErrorMsg)
	} else if len(args) == 0 && (subject == "" || version == "") {
		return errors.New(errors.SchemaOrSubjectErrorMsg)
	}

	return nil
}

func (c *schemaCommand) describe(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return c.describeById(cmd, args)
	}
	return c.describeBySubject(cmd)
}

func (c *schemaCommand) describeById(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}

	schemaID, err := strconv.ParseInt(args[0], 10, 32)
	if err != nil {
		return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.SchemaIntegerErrorMsg, args[0]), errors.SchemaIntegerSuggestions)
	}

	schemaString, _, err := srClient.DefaultApi.GetSchema(ctx, int32(schemaID), nil)
	if err != nil {
		return err
	}

	return c.printSchema(cmd, schemaString.Schema, schemaString.SchemaType, schemaString.References)
}

func (c *schemaCommand) describeBySubject(cmd *cobra.Command) error {
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}
	subject, err := cmd.Flags().GetString("subject")
	if err != nil {
		return err
	}
	version, err := cmd.Flags().GetString("version")
	if err != nil {
		return err
	}
	schemaString, _, err := srClient.DefaultApi.GetSchemaByVersion(ctx, subject, version, nil)
	if err != nil {
		return err
	}
	return c.printSchema(cmd, schemaString.Schema, schemaString.SchemaType, schemaString.References)
}

func (c *schemaCommand) printSchema(cmd *cobra.Command, schema string, sType string, refs []srsdk.SchemaReference) error {
	if sType != "" {
		utils.Println(cmd, "Type: "+sType)
	}
	utils.Println(cmd, "Schema: "+schema)
	if len(refs) > 0 {
		utils.Println(cmd, "References:")
		for i := 0; i < len(refs); i++ {
			utils.Printf(cmd, "\t%s -> %s %d\n", refs[i].Name, refs[i].Subject, refs[i].Version)
		}
	}
	return nil
}
