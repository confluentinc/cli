package schemaregistry

import (
	"context"
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
		Use:         "describe [id]",
		Short:       "Get schema either by schema ID, or by subject/version.",
		Args:        cobra.MaximumNArgs(1),
		PreRunE:     pcmd.NewCLIPreRunnerE(c.preDescribe),
		RunE:        pcmd.NewCLIRunE(c.describe),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
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

	cmd.Flags().String("subject", "", SubjectUsage)
	cmd.Flags().String("version", "", "Version of the schema. Can be a specific version or 'latest'.")
	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)
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
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}
	if len(args) > 0 {
		return c.describeById(cmd, args[0], srClient, ctx)
	}
	return c.describeBySubject(cmd, srClient, ctx)
}

func (c *schemaCommand) describeById(cmd *cobra.Command, id string, srClient *srsdk.APIClient, ctx context.Context) error {
	schemaID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.SchemaIntegerErrorMsg, id), errors.SchemaIntegerSuggestions)
	}

	schemaString, _, err := srClient.DefaultApi.GetSchema(ctx, int32(schemaID), nil)
	if err != nil {
		return err
	}

	return c.printSchema(cmd, schemaID, schemaString.Schema, schemaString.SchemaType, schemaString.References)
}

func (c *schemaCommand) describeBySubject(cmd *cobra.Command, srClient *srsdk.APIClient, ctx context.Context) error {
	subject, err := cmd.Flags().GetString("subject")
	if err != nil {
		return err
	}
	version, err := cmd.Flags().GetString("version")
	if err != nil {
		return err
	}
	schemaString, resp, err := srClient.DefaultApi.GetSchemaByVersion(ctx, subject, version, nil)
	if err != nil {
		return errors.CatchSchemaNotFoundError(err, resp)
	}
	return c.printSchema(cmd, int64(schemaString.Id), schemaString.Schema, schemaString.SchemaType, schemaString.References)
}

func (c *schemaCommand) printSchema(cmd *cobra.Command, schemaID int64, schema string, sType string, refs []srsdk.SchemaReference) error {
	utils.Printf(cmd, "Schema ID: %d\n", schemaID)
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
