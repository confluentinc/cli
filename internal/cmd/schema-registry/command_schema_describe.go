package schemaregistry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/antihax/optional"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/utils"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
)

type schemaOut struct {
	Schemas []srsdk.Schema `json:"schemas"`
}

func (c *schemaCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "describe [id]",
		Short:       "Get schema either by schema ID, or by subject/version.",
		Args:        cobra.MaximumNArgs(1),
		PreRunE:     c.preDescribe,
		RunE:        c.describe,
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
	cmd.Flags().String("version", "", `Version of the schema. Can be a specific version or "latest".`)
	cmd.Flags().Bool("show-references", false, "Display the entire schema graph, including references.")
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
	srClient, ctx, err := getApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}

	showReferences, err := cmd.Flags().GetBool("show-references")
	if err != nil {
		return err
	}

	var id string
	if len(args) == 1 {
		id = args[0]
	}

	if showReferences {
		return describeGraph(cmd, id, srClient, ctx)
	}

	if id != "" {
		return describeById(cmd, id, srClient, ctx)
	}
	return describeBySubject(cmd, srClient, ctx)
}

func describeById(cmd *cobra.Command, id string, srClient *srsdk.APIClient, ctx context.Context) error {
	schemaID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.SchemaIntegerErrorMsg, id), errors.SchemaIntegerSuggestions)
	}

	schemaString, _, err := srClient.DefaultApi.GetSchema(ctx, int32(schemaID), nil)
	if err != nil {
		return err
	}

	return printSchema(cmd, schemaID, schemaString.Schema, schemaString.SchemaType, schemaString.References)
}

func describeBySubject(cmd *cobra.Command, srClient *srsdk.APIClient, ctx context.Context) error {
	subject, err := cmd.Flags().GetString("subject")
	if err != nil {
		return err
	}

	version, err := cmd.Flags().GetString("version")
	if err != nil {
		return err
	}

	schema, httpResp, err := srClient.DefaultApi.GetSchemaByVersion(ctx, subject, version, nil)
	if err != nil {
		return errors.CatchSchemaNotFoundError(err, httpResp)
	}

	return printSchema(cmd, int64(schema.Id), schema.Schema, schema.SchemaType, schema.References)
}

func describeGraph(cmd *cobra.Command, id string, srClient *srsdk.APIClient, ctx context.Context) error {
	subject, err := cmd.Flags().GetString("subject")
	if err != nil {
		return err
	}

	version, err := cmd.Flags().GetString("version")
	if err != nil {
		return err
	}

	visited := make(map[string]bool)
	schemaID := int64(0)
	if len(id) > 0 {
		schemaID, err = strconv.ParseInt(id, 10, 32)
		if err != nil {
			return err
		}
	}

	// A schema graph is a DAG, the root is fetched either by schema id or by subject/version
	// All references are fetched by subject/version
	rootSchema, schemaGraph, err := traverseDAG(srClient, ctx, visited, int32(schemaID), subject, version)
	if err != nil {
		return err
	}

	// Since getting schema by id and by subject/version return different types, i.e., `SchemaString` vs `Schema`,
	// convert root from `SchemaString` to `Schema` so that we only have to deal with a single type, only if the root is fetched by id
	root := convertRootSchema(&rootSchema, int32(schemaID))
	if root != nil {
		schemaGraph = append([]srsdk.Schema{*root}, schemaGraph...)
	}

	b, err := json.Marshal(&schemaOut{schemaGraph})
	if err != nil {
		return err
	}
	utils.Print(cmd, string(pretty.Pretty(b)))

	return nil
}

func traverseDAG(srClient *srsdk.APIClient, ctx context.Context, visited map[string]bool, id int32, subject string, version string) (srsdk.SchemaString, []srsdk.Schema, error) {
	root := srsdk.SchemaString{}
	var schemaGraph []srsdk.Schema
	var refs []srsdk.SchemaReference
	subjectVersionString := strings.Join([]string{subject, version}, "#")

	if id > 0 {
		// should only come here at most once for the root if it is fetched by id
		schemaString, _, err := srClient.DefaultApi.GetSchema(ctx, id, nil)
		if err != nil {
			return srsdk.SchemaString{}, nil, err
		}

		root = schemaString
		refs = schemaString.References
	} else if len(subject) == 0 || len(version) == 0 || visited[subjectVersionString] {
		// dedupe the call if already visited
		return root, schemaGraph, nil
	} else {
		visited[subjectVersionString] = true

		schema, _, err := srClient.DefaultApi.GetSchemaByVersion(ctx, subject, version, &srsdk.GetSchemaByVersionOpts{Deleted: optional.NewBool(true)})
		if err != nil {
			return srsdk.SchemaString{}, nil, err
		}

		schemaGraph = append(schemaGraph, schema)
		refs = schema.References
	}

	for _, reference := range refs {
		_, subGraph, err := traverseDAG(srClient, ctx, visited, 0, reference.Subject, strconv.Itoa(int(reference.Version)))
		if err != nil {
			return srsdk.SchemaString{}, nil, err
		}

		schemaGraph = append(schemaGraph, subGraph...)
	}

	return root, schemaGraph, nil
}

func printSchema(cmd *cobra.Command, schemaID int64, schema string, sType string, refs []srsdk.SchemaReference) error {
	utils.Printf(cmd, "Schema ID: %d\n", schemaID)
	if sType != "" {
		utils.Println(cmd, "Type: "+sType)
	}

	if sType == "PROTOBUF" {
		utils.Println(cmd, "Schema:\n"+schema)
	} else {
		var jsonBuffer bytes.Buffer
		if err := json.Indent(&jsonBuffer, []byte(schema), "", "    "); err != nil {
			return err
		}
		schemaString := jsonBuffer.String()
		utils.Println(cmd, "Schema:\n"+schemaString)
	}

	if len(refs) > 0 {
		utils.Println(cmd, "References:")
		for i := 0; i < len(refs); i++ {
			utils.Printf(cmd, "\t%s -> %s %d\n", refs[i].Name, refs[i].Subject, refs[i].Version)
		}
	}
	return nil
}

func convertRootSchema(root *srsdk.SchemaString, id int32) *srsdk.Schema {
	if len(root.Schema) == 0 {
		return nil
	}

	return &srsdk.Schema{
		Id:         id,
		SchemaType: root.SchemaType,
		References: root.References,
		Schema:     root.Schema,
	}
}
