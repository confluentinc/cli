package schemaregistry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/schemaregistry"
)

type schemaOut struct {
	Schemas []schema `json:"schemas"`
}

type schema struct { // CLI-2918
	Subject    *string                 `json:"subject,omitempty"`
	Version    *int32                  `json:"version,omitempty"`
	Id         *int32                  `json:"id,omitempty"`
	SchemaType *string                 `json:"schemaType,omitempty"`
	References []srsdk.SchemaReference `json:"references,omitempty"`
	Schema     *string                 `json:"schema,omitempty"`
	Metadata   *srsdk.Metadata         `json:"metadata,omitempty"`
	Ruleset    *srsdk.RuleSet          `json:"ruleset,omitempty"`
}

func (c *command) newSchemaDescribeCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe [id]",
		Short: "Get schema by ID, or by subject and version.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.schemaDescribe,
	}

	example1 := examples.Example{
		Text: `Describe the schema with ID "1337".`,
		Code: "confluent schema-registry schema describe 1337",
	}
	example2 := examples.Example{
		Text: `Describe the schema with subject "payments" and version "latest".`,
		Code: "confluent schema-registry schema describe --subject payments --version latest",
	}
	if cfg.IsOnPremLogin() {
		example1.Code += " " + onPremAuthenticationMsg
		example2.Code += " " + onPremAuthenticationMsg
	}
	cmd.Example = examples.BuildExampleString(example1, example2)

	cmd.Flags().String("subject", "", subjectUsage)
	cmd.Flags().String("version", "", `Version of the schema. Can be a specific version or "latest".`)
	cmd.Flags().Bool("show-references", false, "Display the entire schema graph, including references.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		addCaLocationAndClientPathFlags(cmd)
	}
	addSchemaRegistryEndpointFlag(cmd)

	return cmd
}

func (c *command) schemaDescribe(cmd *cobra.Command, args []string) error {
	subject, err := cmd.Flags().GetString("subject")
	if err != nil {
		return err
	}

	version, err := cmd.Flags().GetString("version")
	if err != nil {
		return err
	}

	if len(args) > 0 && (subject != "" || version != "") {
		return fmt.Errorf("cannot specify both schema ID and subject/version")
	} else if len(args) == 0 && (subject == "" || version == "") {
		return fmt.Errorf("must specify either schema ID or subject/version")
	}

	client, err := c.GetSchemaRegistryClient(cmd)
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
		return describeGraph(cmd, id, client)
	}

	if id != "" {
		return describeById(id, client)
	}
	return describeBySubject(cmd, client)
}

func describeById(id string, client *schemaregistry.Client) error {
	schemaId, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return errors.NewErrorWithSuggestions(
			fmt.Sprintf(`invalid schema ID "%s"`, id),
			"Schema ID must be an integer.",
		)
	}

	schemaString, err := client.GetSchema(int32(schemaId), "")
	if err != nil {
		return err
	}

	return printSchema(schemaId, schemaString.GetSchema(), schemaString.GetSchemaType(), schemaString.GetReferences(), schemaString.GetMetadata(), schemaString.GetRuleSet())
}

func describeBySubject(cmd *cobra.Command, client *schemaregistry.Client) error {
	subject, err := cmd.Flags().GetString("subject")
	if err != nil {
		return err
	}

	version, err := cmd.Flags().GetString("version")
	if err != nil {
		return err
	}

	schema, err := client.GetSchemaByVersion(subject, version, false)
	if err != nil {
		return catchSchemaNotFoundError(err, subject, version)
	}

	return printSchema(int64(schema.GetId()), schema.GetSchema(), schema.GetSchemaType(), schema.GetReferences(), schema.GetMetadata(), schema.GetRuleset())
}

func describeGraph(cmd *cobra.Command, id string, client *schemaregistry.Client) error {
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
	if id != "" {
		schemaID, err = strconv.ParseInt(id, 10, 32)
		if err != nil {
			return err
		}
	}

	// A schema graph is a DAG, the root is fetched by ID or by subject and version
	// All references are fetched by subject/version
	rootSchema, schemaGraph, err := traverseDAG(client, visited, int32(schemaID), subject, version)
	if err != nil {
		return err
	}

	// Since getting schema by ID and by subject and version return different types, i.e., `SchemaString` vs `Schema`,
	// convert root from `SchemaString` to `Schema` so that we only have to deal with a single type, only if the root is fetched by id
	root := convertRootSchema(&rootSchema, int32(schemaID))
	if root != nil {
		schemaGraph = append([]schema{*root}, schemaGraph...)
	}

	b, err := json.Marshal(&schemaOut{schemaGraph})
	if err != nil {
		return err
	}
	output.Print(false, string(pretty.Pretty(b)))

	return nil
}

func traverseDAG(client *schemaregistry.Client, visited map[string]bool, id int32, subject, version string) (srsdk.SchemaString, []schema, error) {
	root := srsdk.SchemaString{}
	var schemaGraph []schema
	var refs []srsdk.SchemaReference
	subjectVersionString := strings.Join([]string{subject, version}, "#")

	if id > 0 {
		// should only come here at most once for the root if it is fetched by id
		schemaString, err := client.GetSchema(id, "")
		if err != nil {
			return srsdk.SchemaString{}, nil, err
		}

		root = schemaString
		refs = schemaString.GetReferences()
	} else if subject == "" || version == "" || visited[subjectVersionString] {
		// dedupe the call if already visited
		return root, schemaGraph, nil
	} else {
		visited[subjectVersionString] = true

		srsdkSchema, err := client.GetSchemaByVersion(subject, version, true)
		if err != nil {
			return srsdk.SchemaString{}, nil, err
		}
		schema := schema{
			Subject:    srsdkSchema.Subject,
			Version:    srsdkSchema.Version,
			Id:         srsdkSchema.Id,
			SchemaType: srsdkSchema.SchemaType,
			References: srsdkSchema.GetReferences(),
			Schema:     srsdkSchema.Schema,
			Metadata:   srsdkSchema.Metadata.Get(),
			Ruleset:    srsdkSchema.Ruleset.Get(),
		}

		schemaGraph = append(schemaGraph, schema)
		refs = srsdkSchema.GetReferences()
	}

	for _, reference := range refs {
		_, subGraph, err := traverseDAG(client, visited, 0, reference.GetSubject(), strconv.Itoa(int(reference.GetVersion())))
		if err != nil {
			return srsdk.SchemaString{}, nil, err
		}

		schemaGraph = append(schemaGraph, subGraph...)
	}

	return root, schemaGraph, nil
}

func printSchema(schemaId int64, schema, schemaType string, refs []srsdk.SchemaReference, metadata srsdk.Metadata, ruleset srsdk.RuleSet) error {
	output.Printf(false, "Schema ID: %d\n", schemaId)

	// The backend considers "AVRO" to be the default schema type.
	if schemaType == "" {
		schemaType = "AVRO"
	}
	output.Println(false, "Type: "+schemaType)

	switch schemaType {
	case "JSON", "AVRO":
		var jsonBuffer bytes.Buffer
		if err := json.Indent(&jsonBuffer, []byte(schema), "", "    "); err != nil {
			return err
		}
		schema = jsonBuffer.String()
	}

	output.Println(false, "Schema:")
	output.Println(false, schema)

	if len(refs) > 0 {
		output.Println(false, "References:")
		for i := 0; i < len(refs); i++ {
			output.Printf(false, "\t%s -> %s %d\n", refs[i].GetName(), refs[i].GetSubject(), refs[i].GetVersion())
		}
	}

	if metadata.Properties != nil || metadata.Tags != nil || metadata.Sensitive != nil {
		metadataJson, err := json.Marshal(metadata)
		if err != nil {
			return err
		}
		output.Println(false, "Metadata:")
		output.Println(false, prettyJson(metadataJson))
	}

	if ruleset.DomainRules != nil || ruleset.MigrationRules != nil {
		rulesetJson, err := json.Marshal(ruleset)
		if err != nil {
			return err
		}
		output.Println(false, "Ruleset:")
		output.Println(false, prettyJson(rulesetJson))
	}
	return nil
}

func convertRootSchema(root *srsdk.SchemaString, id int32) *schema {
	if root.GetSchema() == "" {
		return nil
	}

	// The backend considers "AVRO" to be the default schema type.
	if root.GetSchemaType() == "" {
		root.SchemaType = srsdk.PtrString("AVRO")
	}

	return &schema{
		Id:         srsdk.PtrInt32(id),
		SchemaType: root.SchemaType,
		References: root.GetReferences(),
		Schema:     root.Schema,
	}
}
