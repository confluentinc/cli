package schemaregistry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	"github.com/confluentinc/cli/v3/pkg/utils"
)

type RegisterSchemaConfigs struct {
	SchemaDir  string
	Subject    string
	Format     string
	SchemaType string
	SchemaPath string
	Refs       []srsdk.SchemaReference
	Metadata   srsdk.NullableMetadata
	Ruleset    srsdk.NullableRuleSet
	Normalize  bool
}

type RegisterSchemaResponse struct {
	Id int32 `json:"id" yaml:"id"`
}

func ReadSchemaReferences(references string) ([]srsdk.SchemaReference, error) {
	if references == "" {
		return []srsdk.SchemaReference{}, nil
	}

	data, err := os.ReadFile(references)
	if err != nil {
		return nil, err
	}

	var refs []srsdk.SchemaReference
	if err := json.Unmarshal(data, &refs); err != nil {
		return nil, err
	}

	return refs, nil
}

func StoreSchemaReferences(schemaDir string, refs []srsdk.SchemaReference, client *Client) (map[string]string, error) {
	referencePathMap := map[string]string{}
	for _, ref := range refs {
		tempStorePath := filepath.Join(schemaDir, ref.GetName())
		if !utils.FileExists(tempStorePath) {
			schema, err := client.GetSchemaByVersion(ref.GetSubject(), strconv.Itoa(int(ref.GetVersion())), false)
			if err != nil {
				return nil, err
			}
			if err := os.MkdirAll(filepath.Dir(tempStorePath), 0755); err != nil {
				return nil, err
			}
			if err := os.WriteFile(tempStorePath, []byte(schema.GetSchema()), 0644); err != nil {
				return nil, err
			}
		}
		referencePathMap[ref.GetName()] = tempStorePath
	}
	return referencePathMap, nil
}

func RegisterSchemaWithAuth(schemaCfg *RegisterSchemaConfigs, client *Client) (int32, error) {
	schema, err := os.ReadFile(schemaCfg.SchemaPath)
	if err != nil {
		return 0, err
	}

	request := srsdk.RegisterSchemaRequest{
		Schema:     srsdk.PtrString(string(schema)),
		SchemaType: srsdk.PtrString(schemaCfg.SchemaType),
		References: &schemaCfg.Refs,
		Metadata:   schemaCfg.Metadata,
		RuleSet:    schemaCfg.Ruleset,
	}

	response, err := client.Register(schemaCfg.Subject, request, true)
	if err != nil {
		return 0, err
	}

	return response.GetId(), nil
}
