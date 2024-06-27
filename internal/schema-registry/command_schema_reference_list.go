package schemaregistry

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/schemaregistry"
)

type schemaOut struct {
	Schemas []schema `json:"schemas"`
}

func (c *command) newSchemaReferenceListCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [id]",
		Short: "List references by schema ID, or by subject and version.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.schemaReferenceList,
	}

	example1 := examples.Example{
		Text: `List references for the schema with ID "1337".`,
		Code: "confluent schema-registry schema reference list 1337",
	}
	example2 := examples.Example{
		Text: `List references for the schema with subject "payments" and version "latest".`,
		Code: "confluent schema-registry schema reference list --subject payments --version latest",
	}
	if cfg.IsOnPremLogin() {
		example1.Code += " " + onPremAuthenticationMsg
		example2.Code += " " + onPremAuthenticationMsg
	}
	cmd.Example = examples.BuildExampleString(example1, example2)

	cmd.Flags().String("subject", "", subjectUsage)
	cmd.Flags().String("version", "", `Version of the schema. Can be a specific version or "latest".`)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		addCaLocationFlag(cmd)
		addSchemaRegistryEndpointFlag(cmd)
	}

	return cmd
}

func (c *command) schemaReferenceList(cmd *cobra.Command, args []string) error {
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

	var id string
	if len(args) == 1 {
		id = args[0]
	}

	return describeGraph(id, subject, version, client)
}

func describeGraph(id, subject, version string, client *schemaregistry.Client) error {
	visited := make(map[string]bool)
	schemaID := int64(0)
	var err error
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
